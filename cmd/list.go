package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"musiccat/internal/db"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored releases",
	RunE: func(cmd *cobra.Command, args []string) error {
		artist, _ := cmd.Flags().GetString("artist")
		format, _ := cmd.Flags().GetString("format")
		promo, _ := cmd.Flags().GetBool("promo")
		source, _ := cmd.Flags().GetString("source")
		notes, _ := cmd.Flags().GetString("notes")
		sort, _ := cmd.Flags().GetString("sort")
		desc, _ := cmd.Flags().GetBool("desc")

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

		// Query - show individual ownership entries with IDs
		query := `SELECT o.id, r.artist, r.title, r.year, o.format_category, o.format_detail, o.is_promo, o.is_pirate, o.notes
			FROM ownership o
			JOIN releases r ON o.release_id = r.id`
		var queryArgs []interface{}
		where := ""

		if artist != "" {
			where += " r.artist LIKE ?"
			queryArgs = append(queryArgs, "%"+strings.ToLower(artist)+"%")
		}
		if format != "" {
			if where != "" {
				where += " AND"
			}
			where += " o.format_category = ?"
			queryArgs = append(queryArgs, format)
		}
		if promo {
			if where != "" {
				where += " AND"
			}
			where += " o.is_promo = TRUE"
		}
		if source != "" {
			if where != "" {
				where += " AND"
			}
			where += " o.source LIKE ?"
			queryArgs = append(queryArgs, "%"+strings.ToLower(source)+"%")
		}
		if notes != "" {
			if where != "" {
				where += " AND"
			}
			where += " o.notes LIKE ?"
			queryArgs = append(queryArgs, "%"+strings.ToLower(notes)+"%")
		}

		if where != "" {
			query += " WHERE" + where
		}

		// Sorting
		orderClause := " ORDER BY "
		switch sort {
		case "artist":
			orderClause += "r.artist, r.title"
		case "title":
			orderClause += "r.title, r.artist"
		case "year":
			orderClause += "r.year, r.artist, r.title"
		case "format":
			orderClause += "o.format_category, r.artist, r.title"
		case "added":
			orderClause += "o.id"
		default:
			orderClause += "o.id"
		}

		if desc {
			orderClause += " DESC"
		}

		query += orderClause

		rows, err := database.Query(query, queryArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		fmt.Printf("%-3s %-20s %-30s %-6s %-10s %-10s %-5s %-5s %-30s\n", "ID", "Artist", "Title", "Year", "Format", "Detail", "Promo", "Pirate", "Notes")
		fmt.Println(strings.Repeat("-", 125))

		for rows.Next() {
			var id int
			var artist, title, formatCategory string
			var year *int
			var formatDetail, notes *string
			var isPromo, isPirate bool
			var yearNull sql.NullInt32
			var formatDetailNull, notesNull sql.NullString

			err := rows.Scan(&id, &artist, &title, &yearNull, &formatCategory, &formatDetailNull, &isPromo, &isPirate, &notesNull)
			if err != nil {
				return err
			}

			if yearNull.Valid {
				y := int(yearNull.Int32)
				year = &y
			}

			if formatDetailNull.Valid {
				fd := formatDetailNull.String
				formatDetail = &fd
			}

			if notesNull.Valid {
				n := notesNull.String
				notes = &n
			}

			yearStr := ""
			if year != nil {
				yearStr = fmt.Sprintf("%d", *year)
			}

			detailStr := ""
			if formatDetail != nil {
				detailStr = *formatDetail
			}

			promoStr := "no"
			if isPromo {
				promoStr = "yes"
			}

			pirateStr := "no"
			if isPirate {
				pirateStr = "yes"
			}

			notesStr := ""
			if notes != nil {
				notesStr = *notes
			}

			fmt.Printf("%-3d %-20.20s %-30.30s %-6s %-10s %-10.10s %-5s %-5s %-30.30s\n",
				id, artist, title, yearStr, formatCategory, detailStr, promoStr, pirateStr, notesStr)
		}
		return rows.Err()
	},
}

func init() {
	listCmd.Flags().String("artist", "", "Filter by artist (partial, case-insensitive)")
	listCmd.Flags().String("format", "", "Filter by format")
	listCmd.Flags().Bool("promo", false, "Filter promo items only")
	listCmd.Flags().String("source", "", "Filter by source (partial, case-insensitive)")
	listCmd.Flags().String("notes", "", "Filter by notes (partial, case-insensitive)")
	listCmd.Flags().String("sort", "added", "Sort by field (artist, title, year, format, added)")
	listCmd.Flags().Bool("desc", false, "Sort in descending order")
	rootCmd.AddCommand(listCmd)
}
