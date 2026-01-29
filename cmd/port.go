package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"musiccat/internal/db"
	"musiccat/internal/tags"

	"github.com/spf13/cobra"
)

var portCmd = &cobra.Command{
	Use:   "port <search-string> <tag>",
	Short: "Extract notes patterns and convert to tags",
	Long: `Search for an exact string in notes and add a tag to matching ownership records.
Example: musiccat port "Missing sleeve" "missing-sleeve"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		searchString := args[0]
		tagName := args[1]
		
		removeFlag, _ := cmd.Flags().GetBool("remove")

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

		// Find all ownership records with the search string in notes
		query := `SELECT o.id, o.notes, r.artist, r.title 
			FROM ownership o 
			JOIN releases r ON o.release_id = r.id 
			WHERE o.notes LIKE ?`
		
		rows, err := database.Query(query, "%"+searchString+"%")
		if err != nil {
			return err
		}

		// Collect all matching records first (avoid nested queries)
		type match struct {
			ownershipID int
			notes       string
			artist      string
			title       string
		}
		var matches []match
		
		for rows.Next() {
			var ownershipID int
			var notes sql.NullString
			var artist, title string
			
			if err := rows.Scan(&ownershipID, &notes, &artist, &title); err != nil {
				rows.Close()
				return err
			}

			// Skip if notes is null or doesn't actually contain the string (case-sensitive check)
			if !notes.Valid || !strings.Contains(notes.String, searchString) {
				continue
			}

			matches = append(matches, match{
				ownershipID: ownershipID,
				notes:       notes.String,
				artist:      artist,
				title:       title,
			})
		}
		rows.Close()

		if err := rows.Err(); err != nil {
			return err
		}

		// Canonicalize the tag
		canonicalTag := tags.CanonicalizeTag(tagName)
		
		// Get or create the tag (now safe - no open result sets)
		tagID, err := db.GetOrCreateTag(database, canonicalTag)
		if err != nil {
			return err
		}

		// Process all matches
		count := 0
		for _, m := range matches {
			// Add the tag
			if err := db.AddTagToOwnership(database, int64(m.ownershipID), tagID); err != nil {
				return err
			}

			// Optionally remove the string from notes
			if removeFlag {
				newNotes := strings.ReplaceAll(m.notes, searchString, "")
				// Clean up any double spaces or leading/trailing spaces
				newNotes = strings.TrimSpace(strings.Join(strings.Fields(newNotes), " "))
				
				// Update notes (empty string if now empty)
				var notesPtr *string
				if newNotes != "" {
					notesPtr = &newNotes
				}
				
				if err := db.UpdateOwnership(database, int64(m.ownershipID), db.OwnershipUpdate{
					Notes: notesPtr,
				}); err != nil {
					return err
				}
			}

			fmt.Printf("ID %d: %s - %s (tagged: %s)\n", m.ownershipID, m.artist, m.title, canonicalTag)
			count++
		}

		if count == 0 {
			fmt.Printf("No ownership records found with '%s' in notes.\n", searchString)
		} else {
			fmt.Printf("\nTagged %d ownership record(s) with '%s'.\n", count, canonicalTag)
		}

		return nil
	},
}

func init() {
	portCmd.Flags().Bool("remove", false, "Remove the search string from notes after tagging")
	rootCmd.AddCommand(portCmd)
}
