package helpers

import "strings"

// getSortKey returns a sorting key that handles "The" prefixes
func getSortKey(title string) string {
	if strings.HasPrefix(strings.ToLower(title), "the ") {
		return title[4:] // Remove "The " prefix
	}
	return title
}
