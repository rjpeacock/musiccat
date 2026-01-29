package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

type OwnershipRecord struct {
	ID           int
	Artist       string
	Title        string
	Format       string
	PurchaseDate *string
	FormatDetail *string
	Notes        *string
}

var recentCmd = &cobra.Command{
	Use:     "recent",
	Aliases: []string{"r"},
	Short:   "Show recent ownership additions",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		limit, _ := cmd.Flags().GetInt("limit")

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

		// Query
		query := `SELECT o.id, r.artist, r.title, o.format_category, o.acquired_date, o.format_detail, o.notes FROM ownership o JOIN releases r ON o.release_id = r.id`
		var queryArgs []interface{}
		if format != "" {
			query += " WHERE o.format_category = ?"
			queryArgs = append(queryArgs, format)
		}
		query += " ORDER BY o.id DESC LIMIT ?"
		queryArgs = append(queryArgs, limit)

		rows, err := database.Query(query, queryArgs...)
		if err != nil {
			return err
		}

		// Collect all records first to avoid nested queries
		var records []OwnershipRecord
		for rows.Next() {
			var rec OwnershipRecord
			var purchaseDate, formatDetail, notes sql.NullString
			err := rows.Scan(&rec.ID, &rec.Artist, &rec.Title, &rec.Format, &purchaseDate, &formatDetail, &notes)
			if err != nil {
				rows.Close()
				return err
			}
			if purchaseDate.Valid {
				rec.PurchaseDate = &purchaseDate.String
			}
			if formatDetail.Valid {
				rec.FormatDetail = &formatDetail.String
			}
			if notes.Valid {
				rec.Notes = &notes.String
			}
			records = append(records, rec)
		}
		rows.Close()

		if err := rows.Err(); err != nil {
			return err
		}

		// Fetch all tags in one query for efficiency
		tagMap := make(map[int64][]string)
		if len(records) > 0 {
			ownershipIDs := make([]interface{}, len(records))
			for i, rec := range records {
				ownershipIDs[i] = rec.ID
			}

			tagQuery := `SELECT ot.ownership_id, t.name 
				FROM ownership_tags ot 
				JOIN tags t ON ot.tag_id = t.id 
				WHERE ot.ownership_id IN (?` + strings.Repeat(",?", len(ownershipIDs)-1) + `)
				ORDER BY t.name`

			tagRows, err := database.Query(tagQuery, ownershipIDs...)
			if err != nil {
				return err
			}

			for tagRows.Next() {
				var ownershipID int64
				var tagName string
				if err := tagRows.Scan(&ownershipID, &tagName); err != nil {
					tagRows.Close()
					return err
				}
				tagMap[ownershipID] = append(tagMap[ownershipID], tagName)
			}
			tagRows.Close()

			if err := tagRows.Err(); err != nil {
				return err
			}
		}

		fmt.Println("Recent additions:")
		for _, rec := range records {
			fmt.Printf("ID %d: %s - %s (%s", rec.ID, rec.Artist, rec.Title, rec.Format)
			if rec.FormatDetail != nil {
				fmt.Printf(", %s", *rec.FormatDetail)
			}
			if rec.Notes != nil && *rec.Notes != "" {
				// Truncate notes for display but show full content
				notesStr := *rec.Notes
				if len(notesStr) > 40 {
					notesStr = notesStr[:37] + "..."
				}
				fmt.Printf(", notes: %s", notesStr)
			}
			
			// Get tags from map (already fetched in single query)
			tagList := tagMap[int64(rec.ID)]
			if len(tagList) > 0 {
				fmt.Printf(", tags: [%s]", strings.Join(tagList, ", "))
			}
			
			if rec.PurchaseDate != nil {
				fmt.Printf(", purchased %s", *rec.PurchaseDate)
			}
			fmt.Println(")")
		}
		return nil
	},
}

func init() {
	recentCmd.Flags().String("format", "", "Filter by format")
	recentCmd.Flags().Int("limit", 10, "Number of records to show")
	rootCmd.AddCommand(recentCmd)
}
