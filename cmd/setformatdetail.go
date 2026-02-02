package cmd

import (
	"fmt"

	"musiccat/internal/config"

	"github.com/spf13/cobra"
)

var setFormatDetailCmd = &cobra.Command{
	Use:   "set-format-detail <DETAIL>",
	Short: "Set the current batch format detail for additions",
	Long: `Set a default format detail that will be used for all additions until changed or cleared.

Examples:
  musiccat set-format-detail "7\""
  musiccat set-format-detail "12\""
  musiccat set-format-detail "LP"
  musiccat set-format-detail ""  (clears the setting)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		detail := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		cfg.CurrentFormatDetail = detail
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		if detail == "" {
			fmt.Println("Cleared format detail setting")
		} else {
			fmt.Printf("Set default format detail to: %s\n", detail)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setFormatDetailCmd)
}
