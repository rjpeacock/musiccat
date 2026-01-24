package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"musiccat/internal/db"
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
	Use:   "recent",
	Short: "Show recent ownership additions",
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
		query := `SELECT o.id, r.artist, r.title, o.format_category, o.purchase_date, o.format_detail, o.notes FROM ownership o JOIN releases r ON o.release_id = r.id`
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
		defer rows.Close()

		fmt.Println("Recent additions:")
		for rows.Next() {
			var rec OwnershipRecord
			var purchaseDate, formatDetail, notes sql.NullString
			err := rows.Scan(&rec.ID, &rec.Artist, &rec.Title, &rec.Format, &purchaseDate, &formatDetail, &notes)
			if err != nil {
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
			if rec.PurchaseDate != nil {
				fmt.Printf(", purchased %s", *rec.PurchaseDate)
			}
			fmt.Println(")")
		}
		return rows.Err()
	},
}

func init() {
	recentCmd.Flags().String("format", "", "Filter by format")
	recentCmd.Flags().Int("limit", 10, "Number of records to show")
	rootCmd.AddCommand(recentCmd)
}
