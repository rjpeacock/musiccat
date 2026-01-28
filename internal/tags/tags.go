package tags

import (
	"regexp"
	"strings"
)

// CanonicalizeTag converts a tag to its canonical form.
// Rules:
// - Convert to lowercase
// - Trim surrounding whitespace
// - Replace internal whitespace with hyphens
// - Collapse repeated separators
func CanonicalizeTag(tag string) string {
	// Trim whitespace
	tag = strings.TrimSpace(tag)
	
	// Convert to lowercase
	tag = strings.ToLower(tag)
	
	// Replace whitespace with hyphens
	tag = strings.ReplaceAll(tag, " ", "-")
	tag = strings.ReplaceAll(tag, "\t", "-")
	
	// Collapse repeated hyphens
	re := regexp.MustCompile(`-+`)
	tag = re.ReplaceAllString(tag, "-")
	
	// Trim leading/trailing hyphens
	tag = strings.Trim(tag, "-")
	
	return tag
}
