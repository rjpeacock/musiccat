package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"musiccat/internal/db"
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update ownership details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr := args[0]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("invalid ID: %s", idStr)
		}

		manual, _ := cmd.Flags().GetBool("manual-edit")
		cost, _ := cmd.Flags().GetFloat64("cost")
		purchaseDate, _ := cmd.Flags().GetString("purchase-date")
		source, _ := cmd.Flags().GetString("source")
		promo, _ := cmd.Flags().GetBool("promo")
		notes, _ := cmd.Flags().GetString("notes")

		// If no flags provided, default to manual mode
		hasFlags := cmd.Flags().Changed("cost") || cmd.Flags().Changed("purchase-date") ||
			cmd.Flags().Changed("source") || cmd.Flags().Changed("promo") || cmd.Flags().Changed("notes")
		if !hasFlags && !manual {
			manual = true
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

		if manual {
			// Prompt for all fields
			fmt.Println("Enter new values (leave blank to keep current):")

			// Get current values
			var currentArtist, currentTitle string
			var currentYear sql.NullInt32
			var currentMBID sql.NullString
			var currentFormat, currentFormatDetail sql.NullString
			var currentPurchaseDate, currentSource, currentNotes sql.NullString
			var currentCost sql.NullFloat64
			var currentPromo bool
			err = database.QueryRow(`
				SELECT r.artist, r.title, r.year, r.musicbrainz_release_group_id,
					o.format_category, o.format_detail, o.purchase_date, o.cost, o.source, o.notes, o.is_promo
				FROM ownership o
				JOIN releases r ON o.release_id = r.id
				WHERE o.id = ?
			`, id).Scan(&currentArtist, &currentTitle, &currentYear, &currentMBID, &currentFormat, &currentFormatDetail, &currentPurchaseDate, &currentCost, &currentSource, &currentNotes, &currentPromo)
			if err != nil {
				return err
			}

			newArtist := promptOptionalString("Artist", currentArtist)
			newTitle := promptOptionalString("Title", currentTitle)
			newYear := promptOptionalIntUpdate("Year", nil)
			if currentYear.Valid {
				y := int(currentYear.Int32)
				newYear = promptOptionalIntUpdate("Year", &y)
			}
			newMBID := promptOptionalString("MusicBrainz ID", "")
			if currentMBID.Valid {
				newMBID = promptOptionalString("MusicBrainz ID", currentMBID.String)
			}
			if newMBID != nil && *newMBID == "" {
				newMBID = nil
			}
			newFormat := promptValidFormatUpdate("Format category", "")
			if currentFormat.Valid {
				newFormat = promptValidFormatUpdate("Format category", currentFormat.String)
			}
			newFormatDetail := promptOptionalString("Format detail", "")
			if currentFormatDetail.Valid {
				newFormatDetail = promptOptionalString("Format detail", currentFormatDetail.String)
			}
			if newFormatDetail != nil && *newFormatDetail == "" {
				newFormatDetail = nil
			}
			newPurchaseDate := promptOptionalString("Purchase date", "")
			if currentPurchaseDate.Valid {
				newPurchaseDate = promptOptionalString("Purchase date", currentPurchaseDate.String)
			}
			if newPurchaseDate != nil && *newPurchaseDate == "" {
				newPurchaseDate = nil
			}
			newCost := promptOptionalFloat("Cost", nil)
			if currentCost.Valid {
				newCost = promptOptionalFloat("Cost", &currentCost.Float64)
			}
			newSource := promptOptionalString("Source", "")
			if currentSource.Valid {
				newSource = promptOptionalString("Source", currentSource.String)
			}
			if newSource != nil && *newSource == "" {
				newSource = nil
			}
			newNotes := promptOptionalString("Notes", "")
			if currentNotes.Valid {
				newNotes = promptOptionalString("Notes", currentNotes.String)
			}
			if newNotes != nil && *newNotes == "" {
				newNotes = nil
			}
			newPromo := promptOptionalBool("Promo", currentPromo)

			// Update release
			if newArtist != nil || newTitle != nil || newYear != nil || newMBID != nil {
				updateRelease := "UPDATE releases SET"
				var updateArgs []interface{}
				sets := []string{}
				if newArtist != nil {
					sets = append(sets, " artist = ?")
					updateArgs = append(updateArgs, *newArtist)
				}
				if newTitle != nil {
					sets = append(sets, " title = ?")
					updateArgs = append(updateArgs, *newTitle)
				}
				if newYear != nil {
					sets = append(sets, " year = ?")
					updateArgs = append(updateArgs, *newYear)
				}
				if newMBID != nil {
					sets = append(sets, " musicbrainz_release_group_id = ?")
					updateArgs = append(updateArgs, *newMBID)
				}
				if len(sets) > 0 {
					updateRelease += " " + strings.Join(sets, ", ") + " WHERE id = ?"
					updateArgs = append(updateArgs, releaseID)
					_, err = database.Exec(updateRelease, updateArgs...)
					if err != nil {
						return err
					}
				}
			}

			// Update ownership
			updateOwnership := "UPDATE ownership SET"
			var updateArgs []interface{}
			sets := []string{}
			if newFormat != nil {
				sets = append(sets, " format_category = ?")
				updateArgs = append(updateArgs, *newFormat)
			}
			if newFormatDetail != nil {
				sets = append(sets, " format_detail = ?")
				updateArgs = append(updateArgs, *newFormatDetail)
			}
			if newPurchaseDate != nil {
				sets = append(sets, " purchase_date = ?")
				updateArgs = append(updateArgs, *newPurchaseDate)
			}
			if newCost != nil {
				sets = append(sets, " cost = ?")
				updateArgs = append(updateArgs, *newCost)
			}
			if newSource != nil {
				sets = append(sets, " source = ?")
				updateArgs = append(updateArgs, *newSource)
			}
			if newNotes != nil {
				sets = append(sets, " notes = ?")
				updateArgs = append(updateArgs, *newNotes)
			}
			if newPromo != nil {
				sets = append(sets, " is_promo = ?")
				updateArgs = append(updateArgs, *newPromo)
			}
			if len(sets) > 0 {
				updateOwnership += " " + strings.Join(sets, ", ") + " WHERE id = ?"
				updateArgs = append(updateArgs, id)
				_, err = database.Exec(updateOwnership, updateArgs...)
				if err != nil {
					return err
				}
			}
		} else {
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
			if cmd.Flags().Changed("notes") {
				sets = append(sets, " notes = ?")
				updateArgs = append(updateArgs, notes)
			}
			if len(sets) > 0 {
				updateOwnership += " " + strings.Join(sets, ", ") + " WHERE id = ?"
				updateArgs = append(updateArgs, id)
				_, err = database.Exec(updateOwnership, updateArgs...)
				if err != nil {
					return err
				}
			}
		}

		fmt.Println("Updated successfully.")
		return nil
	},
}

func promptOptionalFloat(prompt string, current *float64) *float64 {
	currentStr := ""
	if current != nil {
		currentStr = fmt.Sprintf("%.2f", *current)
	}
	input := promptString(prompt + fmt.Sprintf(" (current: %s): ", currentStr))
	if input == "" {
		return nil
	}
	num, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println("Invalid number, ignoring")
		return nil
	}
	return &num
}

func promptOptionalBool(prompt string, current bool) *bool {
	input := promptString(prompt + fmt.Sprintf(" (current: %t): ", current))
	if input == "" {
		return nil
	}
	if input == "true" || input == "1" || input == "yes" {
		b := true
		return &b
	}
	if input == "false" || input == "0" || input == "no" {
		b := false
		return &b
	}
	fmt.Println("Invalid boolean, ignoring")
	return nil
}

func init() {
	updateCmd.Flags().Bool("manual-edit", false, "Allow editing all fields")
	updateCmd.Flags().Float64("cost", 0, "Update cost")
	updateCmd.Flags().String("purchase-date", "", "Update purchase date")
	updateCmd.Flags().String("source", "", "Update source")
	updateCmd.Flags().Bool("promo", false, "Update promo status")
	updateCmd.Flags().String("notes", "", "Update notes")
	rootCmd.AddCommand(updateCmd)
}
