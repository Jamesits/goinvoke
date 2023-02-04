package utils

import (
	"regexp"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func FormatPublicType(name string) string {
	name = nonAlphanumericRegex.ReplaceAllString(name, "")
	name = strings.ToUpper(string(name[0])) + name[1:]
	return name
}
