package helpers

import (
	"fmt"
	"strings"

	"musiccat/internal/config"
)

var formatDetailSuggestions = map[string][]string{
	"CD":       {"Single", "Album", "EP", "Maxi"},
	"Vinyl":    {"7\"", "10\"", "12\"", "LP"},
	"Cassette": {"Cassette Single", "Cassette Album"},
	"Digital":  {"MP3", "FLAC", "AAC"},
}

func PromptFormatDetail(formatCategory string) string {
	suggestions, exists := formatDetailSuggestions[formatCategory]
	if !exists {
		return PromptString("Format detail (optional): ")
	}

	// Check if there's a saved format detail setting
	cfg, err := config.LoadConfig()
	savedDetail := ""
	if err == nil && cfg.CurrentFormatDetail != "" {
		savedDetail = cfg.CurrentFormatDetail
	}

	suggestionStr := strings.Join(suggestions, ", ")
	
	if savedDetail != "" {
		fmt.Printf("Suggested format details: %s\n", suggestionStr)
		fmt.Printf("Using saved default: %s (press Enter to accept, or type new value)\n", savedDetail)
		input := PromptString(fmt.Sprintf("Format detail [%s]: ", savedDetail))
		if input == "" {
			return savedDetail
		}
		return input
	}

	fmt.Printf("Suggested format details: %s\n", suggestionStr)
	return PromptString("Format detail (optional): ")
}
