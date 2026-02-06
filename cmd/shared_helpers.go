package cmd

import "strings"

// parseArtistName joins multiple command-line arguments into a single artist name.
// This handles multi-word artist names like "Paul Weller" without requiring quotes.
func parseArtistName(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.Join(args, " ")
}
