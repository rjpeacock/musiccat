package helpers

import "fmt"

func PromptOptionalString(prompt, current string) *string {
	input := PromptString(prompt + fmt.Sprintf(" (current: %s): ", current))
	if input == "" {
		return nil
	}
	return &input
}
