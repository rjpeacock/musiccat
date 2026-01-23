package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"musiccat/internal/config"
	"musiccat/internal/db"
)

func main() {
	sqlDB, err := db.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := db.BootstrapDB(sqlDB); err != nil {
		log.Fatal(err)
	}

	rootCmd := &cobra.Command{
		Use:   "musiccat",
		Short: "A CLI for cataloguing music",
		Long:  `musiccat is a tool to manage your personal music collection.`,
	}

	rootCmd.AddCommand(setFormatCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(recentCmd)
	rootCmd.AddCommand(undoCmd)

	rootCmd.Execute()
}

var setFormatCmd = &cobra.Command{
	Use:   "set-format [format]",
	Short: "Set the current batch format",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
		cfg.CurrentFormat = args[0]
		if err := config.SaveConfig(cfg); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Set current format to: %s\n", args[0])
	},
}

var addCmd = &cobra.Command{
	Use:   "add [artist]",
	Short: "Add releases by artist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add placeholder:", args[0])
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored releases",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list placeholder")
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [artist] [title]",
	Short: "Update release information",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("update placeholder: %s - %s\n", args[0], args[1])
	},
}

var recentCmd = &cobra.Command{
	Use:   "recent",
	Short: "Show recent additions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("recent placeholder")
	},
}

var undoCmd = &cobra.Command{
	Use:   "undo [id]",
	Short: "Undo recent additions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("undo placeholder:", args[0])
	},
}
