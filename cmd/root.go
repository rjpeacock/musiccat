package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "musiccat",
	Short: "Catalogue your personal music collection",
}

func Execute() error {
	return rootCmd.Execute()
}
