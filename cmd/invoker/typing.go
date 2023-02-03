package main

import (
	"strings"
)

func formatPublicType(name string) string {
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ToUpper(string(name[0])) + name[1:]
	return name
}
