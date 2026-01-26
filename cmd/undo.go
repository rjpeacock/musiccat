package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"musiccat/internal/db"
)

var undoCmd = &cobra.Command{
	Use:     "undo <ID | all>",
	Aliases: []string{"un"},
	Short:   "Remove ownership entries",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

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

		if target == "all" {
			// Delete the most recent batch (last 10 ownership)
			rows, err := database.Query("SELECT id FROM ownership ORDER BY id DESC LIMIT 10")
			if err != nil {
				return err
			}
			var ids []int
			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					return err
				}
				ids = append(ids, id)
			}
			rows.Close()
			if err := rows.Err(); err != nil {
				return err
			}
			if len(ids) == 0 {
				return fmt.Errorf("no ownership entries to undo")
			}
			if len(ids) > 1 {
				fmt.Printf("This will delete %d ownership entries. Confirm? (y/N): ", len(ids))
				if !confirm() {
					return fmt.Errorf("cancelled")
				}
			}
			for _, id := range ids {
				_, err := database.Exec("DELETE FROM ownership WHERE id = ?", id)
				if err != nil {
					return err
				}
			}
			fmt.Printf("Undid %d ownership entries.\n", len(ids))
		} else {
			// Delete specific ID
			id, err := strconv.Atoi(target)
			if err != nil {
				return fmt.Errorf("invalid ID: %s", target)
			}
			result, err := database.Exec("DELETE FROM ownership WHERE id = ?", id)
			if err != nil {
				return err
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				return fmt.Errorf("no ownership entry with ID %d", id)
			}
			fmt.Printf("Undid ownership entry %d.\n", id)
		}
		return nil
	},
}

func confirm() bool {
	fmt.Print("y/N: ")
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
