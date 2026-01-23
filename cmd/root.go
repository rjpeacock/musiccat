package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "musiccat",
	Short: "Catalogue your personal music collection",
}

var ValidFormats = []string{"CD", "Vinyl", "Tape", "Digital"}

func Execute() error {
	return rootCmd.Execute()
}
