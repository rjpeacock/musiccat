package helpers

import (
	"fmt"
	"strings"
)

var ValidFormats = []string{"CD", "Vinyl", "Cassette", "Digital"}

func PromptValidFormat(prompt string) string {
	for {
		input := PromptString(prompt)
		for _, f := range ValidFormats {
			if strings.EqualFold(input, f) {
				return f
			}
		}
		fmt.Printf("Invalid format. Valid: %s\n", strings.Join(ValidFormats, ", "))
	}
}
