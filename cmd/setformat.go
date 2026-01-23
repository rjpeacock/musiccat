package cmd

import (
	"fmt"
	"strings"

	"musiccat/internal/config"

	"github.com/spf13/cobra"
)

var setFormatCmd = &cobra.Command{
	Use:   "set-format <FORMAT>",
	Short: "Set the current batch format for additions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := strings.ToLower(args[0])
		var format string
		for _, f := range ValidFormats {
			if strings.ToLower(f) == input {
				format = f
				break
			}
		}
		if format == "" {
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

func init() {
	rootCmd.AddCommand(setFormatCmd)
}
