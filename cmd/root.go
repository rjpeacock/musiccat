package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "musiccat",
	Short: "Catalogue your personal music collection",
	Long: `musiccat catalogues your personal music collection.

Format Conventions:
  CD: Album, Single, EP, Maxi-Single, Promo, Digipak, Jewel Case
  Vinyl: LP, 12", 10", 7", Single, EP, Picture Disc, Colored Vinyl
  Cassette: Album, Single, Tape, Cassette

Multiple ownership entries are supported for the same release to handle variants.`,
}

var ValidFormats = []string{"CD", "Vinyl", "Cassette"}

// FormatDetailSuggestions provides recommended format_detail values for each format_category
var formatDetailSuggestions = map[string][]string{
	"CD":       {"Album", "Single", "EP", "Maxi-Single", "Promo", "Digipak", "Jewel Case"},
	"Vinyl":    {"LP", "12\"", "10\"", "7\"", "Single", "EP", "Picture Disc", "Colored Vinyl"},
	"Cassette": {"Album", "Single", "Tape", "Cassette"},
}

func Execute() error {
	return rootCmd.Execute()
}
