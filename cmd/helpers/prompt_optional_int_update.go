package helpers

import (
	"fmt"
	"strconv"
)

func PromptOptionalIntUpdate(prompt string, current *int) *int {
	currentStr := ""
	if current != nil {
		currentStr = strconv.Itoa(*current)
	}
	input := PromptString(prompt + fmt.Sprintf(" (current: %s): ", currentStr))
	if input == "" {
		return nil
	}
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid number, ignoring")
		return nil
	}
	return &num
}
