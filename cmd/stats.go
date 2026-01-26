package cmd

import (
	"fmt"

	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show collection statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Total items owned
		var totalItems int
		err = database.QueryRow("SELECT COUNT(*) FROM ownership").Scan(&totalItems)
		if err != nil {
			return err
		}

		fmt.Printf("Total items owned: %d\n\n", totalItems)

		// Count by format category
		formatCatRows, err := database.Query(`
			SELECT format_category, COUNT(*) 
			FROM ownership 
			GROUP BY format_category 
			ORDER BY COUNT(*) DESC
		`)
		if err != nil {
			return err
		}
		defer formatCatRows.Close()

		fmt.Println("Format category counts:")
		for formatCatRows.Next() {
			var format string
			var count int
			err := formatCatRows.Scan(&format, &count)
			if err != nil {
				return err
			}
			fmt.Printf("  %s: %d\n", format, count)
		}

		// Count by format detail
		formatDetailRows, err := database.Query(`
			SELECT format_detail, COUNT(*) 
			FROM ownership 
			WHERE format_detail IS NOT NULL AND format_detail != ''
			GROUP BY format_detail 
			ORDER BY COUNT(*) DESC
		`)
		if err != nil {
			return err
		}
		defer formatDetailRows.Close()

		fmt.Println("\nFormat detail counts:")
		for formatDetailRows.Next() {
			var detail string
			var count int
			err := formatDetailRows.Scan(&detail, &count)
			if err != nil {
				return err
			}
			fmt.Printf("  %s: %d\n", detail, count)
		}

		// Total spend (sum of cost)
		var totalSpend float64
		err = database.QueryRow("SELECT COALESCE(SUM(cost), 0) FROM ownership WHERE cost IS NOT NULL").Scan(&totalSpend)
		if err != nil {
			return err
		}

		fmt.Printf("\nTotal spend: $%.2f\n", totalSpend)

		// Top 10 artists by items owned
		artistRows, err := database.Query(`
			SELECT r.artist, 
			       COUNT(*) as total,
			       SUM(CASE WHEN o.format_detail = 'Album' THEN 1 ELSE 0 END) as albums,
			       SUM(CASE WHEN o.format_detail = 'Single' THEN 1 ELSE 0 END) as singles
			FROM ownership o
			JOIN releases r ON o.release_id = r.id
			GROUP BY r.artist
			ORDER BY total DESC
			LIMIT 10
		`)
		if err != nil {
			return err
		}
		defer artistRows.Close()

		fmt.Println("\nTop 10 artists by items owned:")
		for artistRows.Next() {
			var artist string
			var total, albums, singles int
			err := artistRows.Scan(&artist, &total, &albums, &singles)
			if err != nil {
				return err
			}
			fmt.Printf("  %s: %d total (%d albums, %d singles)\n", artist, total, albums, singles)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
