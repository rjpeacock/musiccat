package cmd

import (
	"database/sql"
	"fmt"

	"musiccat/internal/db"
	"musiccat/internal/tags"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
	Long:  `Manage global tags: rename or delete tags.`,
}

var tagRenameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a tag",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := args[0]
		newName := args[1]

		// Canonicalize both names
		oldCanonical := tags.CanonicalizeTag(oldName)
		newCanonical := tags.CanonicalizeTag(newName)

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

		// Check if old tag exists
		_, err = db.GetTagIDByName(database, oldCanonical)
		if err == sql.ErrNoRows {
			return fmt.Errorf("tag '%s' does not exist", oldCanonical)
		}
		if err != nil {
			return err
		}

		// Rename the tag
		if err := db.RenameTag(database, oldCanonical, newCanonical); err != nil {
			return err
		}

		fmt.Printf("Renamed tag '%s' to '%s'\n", oldCanonical, newCanonical)
		return nil
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Canonicalize the name
		canonical := tags.CanonicalizeTag(name)

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

		// Check if tag exists
		_, err = db.GetTagIDByName(database, canonical)
		if err == sql.ErrNoRows {
			return fmt.Errorf("tag '%s' does not exist", canonical)
		}
		if err != nil {
			return err
		}

		// Delete the tag
		if err := db.DeleteTag(database, canonical); err != nil {
			return err
		}

		fmt.Printf("Deleted tag '%s'\n", canonical)
		return nil
	},
}

func init() {
	tagCmd.AddCommand(tagRenameCmd)
	tagCmd.AddCommand(tagDeleteCmd)
	rootCmd.AddCommand(tagCmd)
}
