package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update ownership details (flag-driven only)",
	Long: `Update ownership details using flags only. 
Editable fields: purchase_date, cost, source, notes, is_promo.
Non-editable fields: artist, title, year, format, MusicBrainz ID.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr := args[0]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("invalid ID: %s", idStr)
		}

		cost, _ := cmd.Flags().GetFloat64("cost")
		purchaseDate, _ := cmd.Flags().GetString("purchase-date")
		source, _ := cmd.Flags().GetString("source")
		promo, _ := cmd.Flags().GetBool("promo")
		notes, _ := cmd.Flags().GetString("notes")

		// Check if any flags were provided
		hasFlags := cmd.Flags().Changed("cost") || cmd.Flags().Changed("purchase-date") ||
			cmd.Flags().Changed("source") || cmd.Flags().Changed("promo") || cmd.Flags().Changed("notes")
		if !hasFlags {
			return fmt.Errorf("no update flags provided. Use --cost, --purchase-date, --source, --promo, or --notes")
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

		fmt.Printf("Updated ownership ID %d successfully.\n", id)
		return nil
	},
}

func init() {
	updateCmd.Flags().Float64("cost", 0, "Update cost (use 0 to clear)")
	updateCmd.Flags().String("purchase-date", "", "Update purchase date (use empty string to clear)")
	updateCmd.Flags().String("source", "", "Update source (use empty string to clear)")
	updateCmd.Flags().Bool("promo", false, "Update promo status")
	updateCmd.Flags().String("notes", "", "Update notes (use empty string to clear)")
	rootCmd.AddCommand(updateCmd)
}
