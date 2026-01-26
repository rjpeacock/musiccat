package helpers

import (
	"fmt"
	"strconv"
)

func PromptOptionalFloat(prompt string, current float64) *float64 {
	input := PromptString(prompt + fmt.Sprintf(" (current: %.2f): ", current))
	if input == "" {
		return nil
	}
	num, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println("Invalid number, ignoring")
		return nil
	}
	return &num
}
