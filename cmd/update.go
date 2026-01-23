package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"musiccat/internal/db"
)

type UpdateRecord struct {
	OwnershipID  int
	ReleaseID    int
	Artist       string
	Title        string
	Year         *int
	Format       string
	FormatDetail *string
	PurchaseDate *string
	Cost         *float64
	Source       *string
	Notes        *string
	MBID         *string
}

var updateCmd = &cobra.Command{
	Use:   `update "<artist>" "<title>"`,
	Short: "Update release and ownership details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		artist := args[0]
		title := args[1]

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

		// Find matching ownership
		rows, err := database.Query(`
			SELECT o.id, r.id, r.artist, r.title, r.year, o.format_category, o.format_detail, o.purchase_date, o.cost, o.source, o.notes, r.musicbrainz_release_group_id
			FROM ownership o
			JOIN releases r ON o.release_id = r.id
			WHERE r.artist LIKE ? AND r.title LIKE ?
		`, "%"+artist+"%", "%"+title+"%")
		if err != nil {
			return err
		}
		defer rows.Close()

		var records []UpdateRecord
		for rows.Next() {
			var rec UpdateRecord
			var yearNull sql.NullInt32
			var formatDetail, purchaseDate, source, notes, mbid sql.NullString
			var cost sql.NullFloat64
			err := rows.Scan(&rec.OwnershipID, &rec.ReleaseID, &rec.Artist, &rec.Title, &yearNull, &rec.Format, &formatDetail, &purchaseDate, &cost, &source, &notes, &mbid)
			if err != nil {
				return err
			}
			if yearNull.Valid {
				y := int(yearNull.Int32)
				rec.Year = &y
			}
			if formatDetail.Valid {
				rec.FormatDetail = &formatDetail.String
			}
			if purchaseDate.Valid {
				rec.PurchaseDate = &purchaseDate.String
			}
			if cost.Valid {
				rec.Cost = &cost.Float64
			}
			if source.Valid {
				rec.Source = &source.String
			}
			if notes.Valid {
				rec.Notes = &notes.String
			}
			if mbid.Valid {
				rec.MBID = &mbid.String
			}
			records = append(records, rec)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if len(records) == 0 {
			return fmt.Errorf("no matching releases found")
		}

		var selected UpdateRecord
		if len(records) == 1 {
			selected = records[0]
		} else {
			fmt.Println("Multiple matches:")
			for i, rec := range records {
				fmt.Printf("%d. %s - %s (%s", i+1, rec.Artist, rec.Title, rec.Format)
				if rec.Year != nil {
					fmt.Printf(", %d", *rec.Year)
				}
				if rec.FormatDetail != nil {
					fmt.Printf(", %s", *rec.FormatDetail)
				}
				fmt.Println(")")
			}
			idx, err := selectItem("Select which to update (number): ", records)
			if err != nil {
				return err
			}
			selected = idx
		}

		// Prompts for updates
		fmt.Println("Enter new values (leave blank to keep current):")

		newArtist := promptOptionalString("Artist", selected.Artist)
		newTitle := promptOptionalString("Title", selected.Title)
		newYear := promptOptionalIntUpdate("Year", selected.Year)
		newMBID := promptOptionalString("MusicBrainz ID", "")
		if newMBID != nil && *newMBID == "" {
			newMBID = nil
		}
		newFormat := promptValidFormatUpdate("Format category", selected.Format)
		newFormatDetail := promptOptionalString("Format detail", "")
		if newFormatDetail != nil && *newFormatDetail == "" {
			newFormatDetail = nil
		}
		newPurchaseDate := promptOptionalString("Purchase date", "")
		if newPurchaseDate != nil && *newPurchaseDate == "" {
			newPurchaseDate = nil
		}
		newCost := promptOptionalFloat("Cost", selected.Cost)
		newSource := promptOptionalString("Source", "")
		if newSource != nil && *newSource == "" {
			newSource = nil
		}
		newNotes := promptOptionalString("Notes", "")
		if newNotes != nil && *newNotes == "" {
			newNotes = nil
		}

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
				updateRelease += strings.Join(sets, ",")
				updateRelease += " WHERE id = ?"
				updateArgs = append(updateArgs, selected.ReleaseID)
				_, err = database.Exec(updateRelease, updateArgs...)
				if err != nil {
					return err
				}
			}
		}

		// Update ownership
		if newFormat != nil || newFormatDetail != nil || newPurchaseDate != nil || newCost != nil || newSource != nil || newNotes != nil {
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
			if len(sets) > 0 {
				updateOwnership += strings.Join(sets, ",")
				updateOwnership += " WHERE id = ?"
				updateArgs = append(updateArgs, selected.OwnershipID)
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

func promptOptionalString(prompt, current string) *string {
	input := promptString(prompt + fmt.Sprintf(" (current: %s): ", current))
	if input == "" {
		return nil
	}
	return &input
}

func promptOptionalIntUpdate(prompt string, current *int) *int {
	currentStr := ""
	if current != nil {
		currentStr = strconv.Itoa(*current)
	}
	input := promptString(prompt + fmt.Sprintf(" (current: %s): ", currentStr))
	if input == "" {
		return nil
	}
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid number, ignoring")
		return nil
	}
	return &num
}

func promptValidFormatUpdate(prompt, current string) *string {
	input := promptString(prompt + fmt.Sprintf(" (current: %s): ", current))
	if input == "" {
		return nil
	}
	for _, f := range ValidFormats {
		if strings.EqualFold(input, f) {
			return &f
		}
	}
	fmt.Printf("Invalid format. Valid: %s\n", strings.Join(ValidFormats, ", "))
	return nil
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

func init() {
	rootCmd.AddCommand(updateCmd)
}
