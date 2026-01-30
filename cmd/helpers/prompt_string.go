package helpers

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func PromptString(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
