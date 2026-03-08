package git

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ExtractWorkItemID extracts a work item ID from a branch name using the given pattern.
// Default pattern matches digits after a slash, e.g. "feature/1234-description" -> "1234".
func ExtractWorkItemID(branch, pattern string) string {
	if pattern == "" {
		pattern = `(?:/)\d+`
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fallback: find first sequence of digits after a /
		re = regexp.MustCompile(`/(\d+)`)
		m := re.FindStringSubmatch(branch)
		if len(m) >= 2 {
			return m[1]
		}
		return ""
	}
	match := re.FindString(branch)
	// Strip leading /
	match = strings.TrimLeft(match, "/")
	return match
}

// BuildBranchName creates a branch name from a prefix, work item ID, and title.
// e.g. ("feature", 1234, "Fix login timeout") -> "feature/1234-fix-login-timeout"
func BuildBranchName(prefix string, wiID int, title string) string {
	slug := slugify(title)
	if slug == "" {
		return fmt.Sprintf("%s/%d", prefix, wiID)
	}
	// Limit slug length to keep branch names reasonable
	if len(slug) > 50 {
		slug = slug[:50]
		// Don't end on a dash
		slug = strings.TrimRight(slug, "-")
	}
	return fmt.Sprintf("%s/%d-%s", prefix, wiID, slug)
}

// slugify converts a title to a URL-safe lowercase slug.
func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.TrimRight(b.String(), "-")
}
