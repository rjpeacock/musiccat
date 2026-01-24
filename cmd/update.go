package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update ownership details (interactive by default)",
	Long: `Update ownership details interactively or using flags.
Editable fields: purchase_date, cost, source, notes, is_promo, is_pirate.
Non-editable fields: artist, title, year, format, MusicBrainz ID.
Interactive mode is used when no flags are provided.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr := args[0]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("invalid ID: %s", idStr)
		}

		cost, _ := cmd.Flags().GetFloat64("cost")
		purchaseDate, _ := cmd.Flags().GetString("purchase-date")
		source, _ := cmd.Flags().GetString("source")
		promo, _ := cmd.Flags().GetBool("promo")
		pirate, _ := cmd.Flags().GetBool("pirate")
		notes, _ := cmd.Flags().GetString("notes")
		formatDetail, _ := cmd.Flags().GetString("format-detail")

		// Check if any flags were provided
		hasFlags := cmd.Flags().Changed("cost") || cmd.Flags().Changed("purchase-date") ||
			cmd.Flags().Changed("source") || cmd.Flags().Changed("promo") || cmd.Flags().Changed("pirate") || cmd.Flags().Changed("notes") || cmd.Flags().Changed("format-detail")
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
		if cmd.Flags().Changed("purchase-date") {
			sets = append(sets, " purchase_date = ?")
			updateArgs = append(updateArgs, purchaseDate)
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

		fmt.Printf("Updated ownership ID %d successfully.\n", id)
		return nil
	},
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

	// Get current values
	var currentNotes, currentSource, currentFormatDetail, currentPurchaseDate string
	var currentCost float64
	var currentPromo, currentPirate bool

	err = database.QueryRow(`
		SELECT notes, source, format_detail, purchase_date, cost, is_promo, is_pirate 
		FROM ownership WHERE id = ?`, id).Scan(
		&currentNotes, &currentSource, &currentFormatDetail, &currentPurchaseDate,
		&currentCost, &currentPromo, &currentPirate)
	if err != nil {
		return fmt.Errorf("ownership with ID %d not found", id)
	}

	// Display current values and prompt for new ones
	fmt.Printf("Current values for ownership ID %d:\n", id)
	fmt.Printf("Notes: %s\n", safeString(currentNotes))
	fmt.Printf("Source: %s\n", safeString(currentSource))
	fmt.Printf("Format Detail: %s\n", safeString(currentFormatDetail))
	fmt.Printf("Purchase Date: %s\n", safeString(currentPurchaseDate))
	fmt.Printf("Cost: %.2f\n", currentCost)
	fmt.Printf("Promo: %t\n", currentPromo)
	fmt.Printf("Pirate: %t\n", currentPirate)
	fmt.Println()

	// Prompt for updates
	newNotes := promptOptionalString("Notes", currentNotes)
	newSource := promptOptionalString("Source", currentSource)
	newFormatDetail := promptOptionalString("Format detail", currentFormatDetail)
	newPurchaseDate := promptOptionalString("Purchase date", currentPurchaseDate)
	newCost := promptOptionalFloat("Cost", currentCost)
	newPromo := promptOptionalBool("Promo", currentPromo)
	newPirate := promptOptionalBool("Pirate", currentPirate)

	// Apply updates using the new UpdateOwnership function
	return db.UpdateOwnership(database, int64(id), newFormatDetail, newPurchaseDate, newCost, newSource, newNotes, nil, newPromo, newPirate)
}

func safeString(s string) string {
	if s == "" {
		return "(empty)"
	}
	return s
}

func init() {
	updateCmd.Flags().Float64("cost", 0, "Update cost (use 0 to clear)")
	updateCmd.Flags().String("purchase-date", "", "Update purchase date (use empty string to clear)")
	updateCmd.Flags().String("source", "", "Update source (use empty string to clear)")
	updateCmd.Flags().Bool("promo", false, "Update promo status")
	updateCmd.Flags().Bool("pirate", false, "Update pirate status")
	updateCmd.Flags().String("notes", "", "Update notes (use empty string to clear)")
	updateCmd.Flags().String("format-detail", "", "Update format detail (use empty string to clear)")
	rootCmd.AddCommand(updateCmd)
}
