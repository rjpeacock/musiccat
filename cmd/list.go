package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"musiccat/cmd/helpers"
	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list [artist]",
	Aliases: []string{"ls", "l"},
	Short:   "List all stored releases",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		artist, _ := cmd.Flags().GetString("artist")
		
		// If positional argument provided, use it as artist filter
		if len(args) > 0 {
			artist = args[0]
		}
		
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
		query := `SELECT o.id, r.artist, r.title, r.year, o.format_category, o.format_detail, o.is_promo, o.is_pirate, o.acquired_date, o.notes
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

		// Collect all records first to avoid nested queries
		type record struct {
			id            int
			artist        string
			title         string
			year          *int
			formatCategory string
			formatDetail  *string
			isPromo       bool
			isPirate      bool
			acquiredDate  *string
			notes         *string
		}
		var records []record

		for rows.Next() {
			var id int
			var artist, title, formatCategory string
			var yearNull sql.NullInt32
			var formatDetailNull, acquiredDateNull, notesNull sql.NullString
			var isPromo, isPirate bool

			err := rows.Scan(&id, &artist, &title, &yearNull, &formatCategory, &formatDetailNull, &isPromo, &isPirate, &acquiredDateNull, &notesNull)
			if err != nil {
				rows.Close()
				return err
			}

			rec := record{
				id:             id,
				artist:         artist,
				title:          title,
				formatCategory: formatCategory,
				isPromo:        isPromo,
				isPirate:       isPirate,
			}

			if yearNull.Valid {
				y := int(yearNull.Int32)
				rec.year = &y
			}
			if formatDetailNull.Valid {
				fd := formatDetailNull.String
				rec.formatDetail = &fd
			}
			if acquiredDateNull.Valid {
				ad := acquiredDateNull.String
				rec.acquiredDate = &ad
			}
			if notesNull.Valid {
				n := notesNull.String
				rec.notes = &n
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
				ownershipIDs[i] = rec.id
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

		fmt.Printf("%-5s %-25s %-40s %-6s %-10s %-10s %-5s %-5s %-12s %-10s %-30s %-20s\n", "ID", "Artist", "Title", "Year", "Format", "Detail", "Promo", "Pirate", "Acquired", "Importance", "Notes", "Tags")
		fmt.Println(strings.Repeat("-", 187))

		// Now process and display all records
		for _, rec := range records {
			yearStr := ""
			if rec.year != nil {
				yearStr = fmt.Sprintf("%d", *rec.year)
			}

			detailStr := ""
			if rec.formatDetail != nil {
				detailStr = *rec.formatDetail
			}

			acquiredStr := ""
			if rec.acquiredDate != nil {
				acquiredStr = *rec.acquiredDate
			}

			promoStr := "no"
			if rec.isPromo {
				promoStr = "yes"
			}

			pirateStr := "no"
			if rec.isPirate {
				pirateStr = "yes"
			}

			notesStr := ""
			if rec.notes != nil {
				notesStr = *rec.notes
			}

			importance := helpers.DeriveImportance(rec.isPirate, rec.isPromo, rec.formatDetail)
			importanceStr := strings.Repeat("*", importance)

			// Get tags from map (already fetched in single query)
			tagList := tagMap[int64(rec.id)]
			tagsStr := ""
			if len(tagList) > 0 {
				tagsStr = "[" + strings.Join(tagList, ", ") + "]"
			}

			fmt.Printf("%-5d %-25.25s %-40.40s %-6s %-10s %-10.10s %-5s %-5s %-12.12s %-10s %-30.30s %-20.20s\n",
				rec.id, rec.artist, rec.title, yearStr, rec.formatCategory, detailStr, promoStr, pirateStr, acquiredStr, importanceStr, notesStr, tagsStr)
		}
		return nil
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
	
	// Shell completions
	listCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.ValidFormats, cobra.ShellCompDirectiveNoFileComp
	})
	listCmd.RegisterFlagCompletionFunc("sort", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"artist", "title", "year", "format", "added"}, cobra.ShellCompDirectiveNoFileComp
	})
	
	rootCmd.AddCommand(listCmd)
}
