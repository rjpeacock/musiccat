package helpers

import (
	"fmt"
	"strings"
)

func PromptOptionalBool(prompt string, current bool) *bool {
	currentStr := "no"
	if current {
		currentStr = "yes"
	}
	input := PromptString(prompt + fmt.Sprintf(" (current: %s): ", currentStr))
	if input == "" {
		return nil
	}
	value := strings.EqualFold(input, "yes") || strings.EqualFold(input, "y") || strings.EqualFold(input, "true")
	return &value
}
