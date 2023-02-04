package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

// FormatPublicType returns a human-readable Golang public type name from an arbitrary non-empty string.
func FormatPublicType(name string) string {
	name = nonAlphanumericRegex.ReplaceAllString(name, "")

	if unicode.IsDigit(rune(name[0])) {
		name = "T" + name
	} else {
		name = strings.ToUpper(string(name[0])) + name[1:]
	}

	return name
}
