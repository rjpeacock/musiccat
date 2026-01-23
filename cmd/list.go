package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"musiccat/internal/db"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored releases",
	RunE: func(cmd *cobra.Command, args []string) error {
		artist, _ := cmd.Flags().GetString("artist")
		format, _ := cmd.Flags().GetString("format")

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
		query := `SELECT r.id, r.artist, r.title, r.year, GROUP_CONCAT(o.format_category || CASE WHEN o.format_detail IS NOT NULL THEN ' (' || o.format_detail || ')' ELSE '' END) as formats
			FROM releases r
			LEFT JOIN ownership o ON r.id = o.release_id`
		var queryArgs []interface{}
		where := ""
		if artist != "" {
			where += " r.artist LIKE ?"
			queryArgs = append(queryArgs, "%"+artist+"%")
		}
		if format != "" {
			if where != "" {
				where += " AND"
			}
			where += " EXISTS (SELECT 1 FROM ownership o2 WHERE o2.release_id = r.id AND o2.format_category = ?)"
			queryArgs = append(queryArgs, format)
		}
		if where != "" {
			query += " WHERE" + where
		}
		query += " GROUP BY r.id, r.artist, r.title, r.year ORDER BY r.artist, r.title"

		rows, err := database.Query(query, queryArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		fmt.Println("Stored releases:")
		for rows.Next() {
			var id int
			var artist, title string
			var year *int
			var formats string
			var yearNull sql.NullInt32
			err := rows.Scan(&id, &artist, &title, &yearNull, &formats)
			if err != nil {
				return err
			}
			if yearNull.Valid {
				y := int(yearNull.Int32)
				year = &y
			}
			fmt.Printf("%s - %s", artist, title)
			if year != nil {
				fmt.Printf(" (%d)", *year)
			}
			if formats != "" {
				fmt.Printf(" [%s]", formats)
			}
			fmt.Println()
		}
		return rows.Err()
	},
}

func init() {
	listCmd.Flags().String("artist", "", "Filter by artist")
	listCmd.Flags().String("format", "", "Filter by format")
	rootCmd.AddCommand(listCmd)
}
