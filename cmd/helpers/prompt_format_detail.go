package helpers

import (
	"fmt"
	"strings"
)

var formatDetailSuggestions = map[string][]string{
	"CD":       {"CD Single", "CD Album", "CD EP", "CD Maxi"},
	"Vinyl":    {"7\"", "10\"", "12\"", "LP"},
	"Cassette": {"Cassette Single", "Cassette Album"},
	"Digital":  {"MP3", "FLAC", "AAC"},
}

func PromptFormatDetail(formatCategory string) string {
	suggestions, exists := formatDetailSuggestions[formatCategory]
	if !exists {
		return PromptString("Format detail (optional): ")
	}

	suggestionStr := strings.Join(suggestions, ", ")
	fmt.Printf("Suggested format details: %s\n", suggestionStr)
	return PromptString("Format detail (optional): ")
}
