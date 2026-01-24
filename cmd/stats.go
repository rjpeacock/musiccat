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

		// Count by format
		formatRows, err := database.Query(`
			SELECT format_category, COUNT(*) 
			FROM ownership 
			GROUP BY format_category 
			ORDER BY COUNT(*) DESC
		`)
		if err != nil {
			return err
		}
		defer formatRows.Close()

		fmt.Println("Format counts:")
		for formatRows.Next() {
			var format string
			var count int
			err := formatRows.Scan(&format, &count)
			if err != nil {
				return err
			}
			fmt.Printf("  %s: %d\n", format, count)
		}

		// Promo vs non-promo count
		var promoCount int
		err = database.QueryRow("SELECT COUNT(*) FROM ownership WHERE is_promo = TRUE").Scan(&promoCount)
		if err != nil {
			return err
		}
		nonPromoCount := totalItems - promoCount

		fmt.Printf("\nPromo items: %d\n", promoCount)
		fmt.Printf("Non-promo items: %d\n", nonPromoCount)

		// Total spend (sum of cost)
		var totalSpend float64
		err = database.QueryRow("SELECT COALESCE(SUM(cost), 0) FROM ownership WHERE cost IS NOT NULL").Scan(&totalSpend)
		if err != nil {
			return err
		}

		fmt.Printf("\nTotal spend: $%.2f\n", totalSpend)
		fmt.Printf("\nTotal items owned: %d\n", totalItems)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
