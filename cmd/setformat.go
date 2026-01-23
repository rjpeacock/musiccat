package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"musiccat/internal/config"
)

var setFormatCmd = &cobra.Command{
	Use:   "set-format <FORMAT>",
	Short: "Set the current batch format for additions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format := strings.Title(strings.ToLower(args[0]))
		if !isValidFormat(format) {
			return fmt.Errorf("invalid format: %s. Valid formats: %s", args[0], strings.Join(ValidFormats, ", "))
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		cfg.CurrentFormat = format
		return config.SaveConfig(cfg)
	},
}

func isValidFormat(format string) bool {
	for _, f := range ValidFormats {
		if f == format {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(setFormatCmd)
}
