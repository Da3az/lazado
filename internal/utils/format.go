package utils

import (
	"regexp"
	"strings"
)

// Truncate shortens a string to max length, adding "..." if truncated.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// StripHTML removes HTML tags from a string (for work item descriptions).
func StripHTML(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	clean := re.ReplaceAllString(s, "")
	// Collapse multiple whitespace
	clean = regexp.MustCompile(`\s+`).ReplaceAllString(clean, " ")
	return strings.TrimSpace(clean)
}

// PadRight pads a string to the given width with spaces.
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}
