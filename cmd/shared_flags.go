package cmd

import "github.com/spf13/cobra"

// ReleaseGroupFilterFlags adds common release group filtering flags to a command
func ReleaseGroupFilterFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("exact", false, "Exact artist name match")
	cmd.Flags().Bool("album", false, "Show only albums (excludes compilations, live, soundtracks)")
	cmd.Flags().Bool("single", false, "Show only singles")
	cmd.Flags().Bool("ep", false, "Show only EPs")
	cmd.Flags().Bool("compilation", false, "Show only compilations")
	cmd.Flags().Bool("live", false, "Show only live albums")
	cmd.Flags().Bool("soundtrack", false, "Show only soundtracks")
	cmd.Flags().Int("year", 0, "Filter by release year")
	cmd.Flags().String("title", "", "Filter by release title (partial match)")
}
