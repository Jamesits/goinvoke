package utils

import (
	"regexp"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// FormatPublicType returns a human-readable Golang public type name from an arbitrary string
func FormatPublicType(name string) string {
	name = nonAlphanumericRegex.ReplaceAllString(name, "")
	name = strings.ToUpper(string(name[0])) + name[1:]
	return name
}
