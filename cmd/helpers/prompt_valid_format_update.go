package helpers

import (
	"fmt"
	"strings"
)

func PromptValidFormatUpdate(prompt, current string) *string {
	input := PromptString(prompt + fmt.Sprintf(" (current: %s): ", current))
	if input == "" {
		return nil
	}
	for _, f := range ValidFormats {
		if strings.EqualFold(input, f) {
			return &f
		}
	}
	fmt.Printf("Invalid format. Valid: %s\n", strings.Join(ValidFormats, ", "))
	return nil
}
