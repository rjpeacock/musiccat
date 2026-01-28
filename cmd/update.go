package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"musiccat/cmd/helpers"
	"musiccat/internal/db"
	"musiccat/internal/tags"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update [id]",
	Aliases: []string{"u", "up"},
	Short:   "Update ownership details (interactive by default)",
	Long: `Update ownership details interactively or using flags.
Editable fields: acquired_date, cost, source, notes, is_promo, is_pirate.
Non-editable fields: artist, title, year, format, MusicBrainz ID.
Interactive mode is used when no flags are provided.

If no ID is provided, updates the most recently added item.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int
		var err error
		
		if len(args) == 0 {
			// No ID provided - get the most recent ownership ID
			database, err := db.OpenDB()
			if err != nil {
				return err
			}
			defer database.Close()
			
			if err := db.BootstrapDB(database); err != nil {
				return err
			}
			
			err = database.QueryRow("SELECT id FROM ownership ORDER BY id DESC LIMIT 1").Scan(&id)
			if err != nil {
				return fmt.Errorf("no items found in collection")
			}
		} else {
			idStr := args[0]
			id, err = strconv.Atoi(idStr)
			if err != nil {
				return fmt.Errorf("invalid ID: %s", idStr)
			}
		}

		cost, _ := cmd.Flags().GetFloat64("cost")
		acquiredDate, _ := cmd.Flags().GetString("acquired-date")
		source, _ := cmd.Flags().GetString("source")
		promo, _ := cmd.Flags().GetBool("promo")
		pirate, _ := cmd.Flags().GetBool("pirate")
		notes, _ := cmd.Flags().GetString("notes")
		formatDetail, _ := cmd.Flags().GetString("format-detail")
		tag, _ := cmd.Flags().GetString("tag")
		removeTag, _ := cmd.Flags().GetString("remove-tag")
		setTag, _ := cmd.Flags().GetString("set-tag")

		// Check if any flags were provided
		hasFlags := cmd.Flags().Changed("cost") || cmd.Flags().Changed("acquired-date") ||
			cmd.Flags().Changed("source") || cmd.Flags().Changed("promo") || cmd.Flags().Changed("pirate") || 
			cmd.Flags().Changed("notes") || cmd.Flags().Changed("format-detail") ||
			cmd.Flags().Changed("tag") || cmd.Flags().Changed("remove-tag") || cmd.Flags().Changed("set-tag")
		if !hasFlags {
			// Default to interactive mode if no flags provided
			return updateInteractive(id)
		}

		// Open DB
		database, err := db.OpenDB()
		if err != nil {
			return err
		}
		defer database.Close()

		// Bootstrap if needed
		if err := db.BootstrapDB(database); err != nil {
			return err
		}

		// Check if ownership exists
		var releaseID int
		err = database.QueryRow("SELECT release_id FROM ownership WHERE id = ?", id).Scan(&releaseID)
		if err != nil {
			return fmt.Errorf("ownership with ID %d not found", id)
		}

		// Update with flags
		updateOwnership := "UPDATE ownership SET"
		var updateArgs []interface{}
		sets := []string{}

		if cmd.Flags().Changed("cost") {
			sets = append(sets, " cost = ?")
			updateArgs = append(updateArgs, cost)
		}
		if cmd.Flags().Changed("acquired-date") {
			sets = append(sets, " acquired_date = ?")
			updateArgs = append(updateArgs, acquiredDate)
		}
		if cmd.Flags().Changed("source") {
			sets = append(sets, " source = ?")
			updateArgs = append(updateArgs, source)
		}
		if cmd.Flags().Changed("promo") {
			sets = append(sets, " is_promo = ?")
			updateArgs = append(updateArgs, promo)
		}
		if cmd.Flags().Changed("pirate") {
			sets = append(sets, " is_pirate = ?")
			updateArgs = append(updateArgs, pirate)
		}
		if cmd.Flags().Changed("notes") {
			sets = append(sets, " notes = ?")
			updateArgs = append(updateArgs, notes)
		}
		if cmd.Flags().Changed("format-detail") {
			sets = append(sets, " format_detail = ?")
			updateArgs = append(updateArgs, formatDetail)
		}

		if len(sets) > 0 {
			updateOwnership += " " + strings.Join(sets, ", ") + " WHERE id = ?"
			updateArgs = append(updateArgs, id)
			_, err = database.Exec(updateOwnership, updateArgs...)
			if err != nil {
				return err
			}
		}

		// Handle tag operations
		if cmd.Flags().Changed("set-tag") {
			// Replace all tags
			if err := db.RemoveAllTagsFromOwnership(database, int64(id)); err != nil {
				return err
			}
			if setTag != "" {
				if err := addTagsToOwnership(database, int64(id), setTag); err != nil {
					return err
				}
			}
		} else {
			// Add tags if specified
			if cmd.Flags().Changed("tag") && tag != "" {
				if err := addTagsToOwnership(database, int64(id), tag); err != nil {
					return err
				}
			}
			
			// Remove tags if specified
			if cmd.Flags().Changed("remove-tag") && removeTag != "" {
				if err := removeTagsFromOwnership(database, int64(id), removeTag); err != nil {
					return err
				}
			}
		}

		fmt.Printf("Updated ownership ID %d successfully.\n", id)
		return nil
	},
}

// addTagsToOwnership adds comma-separated tags to an ownership record.
func addTagsToOwnership(database *sql.DB, ownershipID int64, tagStr string) error {
	tagNames := strings.Split(tagStr, ",")
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		
		// Canonicalize the tag name
		canonical := tags.CanonicalizeTag(name)
		
		// Get or create the tag
		tagID, err := db.GetOrCreateTag(database, canonical)
		if err != nil {
			return err
		}
		
		// Add the tag to the ownership
		if err := db.AddTagToOwnership(database, ownershipID, tagID); err != nil {
			return err
		}
	}
	return nil
}

// removeTagsFromOwnership removes comma-separated tags from an ownership record.
func removeTagsFromOwnership(database *sql.DB, ownershipID int64, tagStr string) error {
	tagNames := strings.Split(tagStr, ",")
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		
		// Canonicalize the tag name
		canonical := tags.CanonicalizeTag(name)
		
		// Get the tag ID
		tagID, err := db.GetTagIDByName(database, canonical)
		if err == sql.ErrNoRows {
			// Tag doesn't exist, skip silently
			continue
		}
		if err != nil {
			return err
		}
		
		// Remove the tag from the ownership
		if err := db.RemoveTagFromOwnership(database, ownershipID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func updateInteractive(id int) error {
	// Open DB
	database, err := db.OpenDB()
	if err != nil {
		return err
	}
	defer database.Close()

	// Bootstrap if needed
	if err := db.BootstrapDB(database); err != nil {
		return err
	}

	// Get current values including release information
	var currentNotes, currentSource, currentFormatDetail, currentAcquiredDate sql.NullString
	var currentCost sql.NullFloat64
	var currentPromo, currentPirate bool
	var artist, title, formatCategory string
	var yearNull sql.NullInt32

	err = database.QueryRow(`
		SELECT r.artist, r.title, r.year, o.format_category, o.notes, o.source, o.format_detail, o.acquired_date, o.cost, o.is_promo, o.is_pirate 
		FROM ownership o 
		JOIN releases r ON o.release_id = r.id 
		WHERE o.id = ?`, id).Scan(
		&artist, &title, &yearNull, &formatCategory,
		&currentNotes, &currentSource, &currentFormatDetail, &currentAcquiredDate,
		&currentCost, &currentPromo, &currentPirate)
	if err != nil {
		return fmt.Errorf("ownership with ID %d not found", id)
	}

	// Display release information
	yearStr := ""
	if yearNull.Valid {
		yearStr = fmt.Sprintf("%d", yearNull.Int32)
	}
	fmt.Printf("=== Editing ownership ID %d ===\n", id)
	fmt.Printf("%s - %s (%s) [%s]\n\n", artist, title, yearStr, formatCategory)

	// Display current values and prompt for new ones
	fmt.Println("Current values:")
	fmt.Printf("Notes: %s\n", safeNullString(currentNotes))
	fmt.Printf("Source: %s\n", safeNullString(currentSource))
	fmt.Printf("Format Detail: %s\n", safeNullString(currentFormatDetail))
	fmt.Printf("Acquired Date: %s\n", safeNullString(currentAcquiredDate))
	if currentCost.Valid {
		fmt.Printf("Cost: %.2f\n", currentCost.Float64)
	} else {
		fmt.Printf("Cost: (empty)\n")
	}
	fmt.Printf("Promo: %t\n", currentPromo)
	fmt.Printf("Pirate: %t\n", currentPirate)
	fmt.Println()

	// Prompt for updates
	newNotes := helpers.PromptOptionalString("Notes", nullStringValue(currentNotes))
	newSource := helpers.PromptOptionalString("Source", nullStringValue(currentSource))
	newFormatDetail := helpers.PromptOptionalString("Format detail", nullStringValue(currentFormatDetail))
	newAcquiredDate := helpers.PromptOptionalString("Acquired date", nullStringValue(currentAcquiredDate))
	newCost := helpers.PromptOptionalFloat("Cost", nullFloat64Value(currentCost))
	newPromo := helpers.PromptOptionalBool("Promo", currentPromo)
	newPirate := helpers.PromptOptionalBool("Pirate", currentPirate)

	// Apply updates using the new UpdateOwnership function
	return db.UpdateOwnership(database, int64(id), newFormatDetail, newAcquiredDate, newCost, newSource, newNotes, nil, newPromo, newPirate)
}

func safeNullString(ns sql.NullString) string {
	if ns.Valid {
		if ns.String == "" {
			return "(empty)"
		}
		return ns.String
	}
	return "(empty)"
}

func nullStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullFloat64Value(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

func init() {
	updateCmd.Flags().Float64("cost", 0, "Update cost (use 0 to clear)")
	updateCmd.Flags().String("acquired-date", "", "Update acquired date (use empty string to clear)")
	updateCmd.Flags().String("source", "", "Update source (use empty string to clear)")
	updateCmd.Flags().Bool("promo", false, "Update promo status")
	updateCmd.Flags().Bool("pirate", false, "Update pirate status")
	updateCmd.Flags().String("notes", "", "Update notes (use empty string to clear)")
	updateCmd.Flags().String("format-detail", "", "Update format detail (use empty string to clear)")
	updateCmd.Flags().String("tag", "", "Add tags (comma-separated)")
	updateCmd.Flags().String("remove-tag", "", "Remove tags (comma-separated)")
	updateCmd.Flags().String("set-tag", "", "Replace all tags (comma-separated)")
	rootCmd.AddCommand(updateCmd)
}
