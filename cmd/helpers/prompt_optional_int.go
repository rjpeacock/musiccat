package helpers

import (
	"fmt"
	"strconv"
)

func PromptOptionalInt(prompt string) *int {
	input := PromptString(prompt)
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
