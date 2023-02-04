package benchmark

import (
	"regexp"
	"strings"
	"testing"
	"unicode"
)

var testSet = []string{
	"abcdefg",
	"Abcdefg",
	"te st ـا ą ٦测试文字あいう__*",
	"te����st",
	"\tr1JhRgg11TQ2neeKXzCb\n1ZEh0qBDKeeYu3qxSGiG\nXRoUAFavpLSVDbQThYw1\nI9MS2f8ChzuZG3y4Lbe9\nzgrc4PlkWePY5fY7vWRZ\nI75l9weN8uJl51v8YHej\nQ5K1KdBEGEdhB9Zu45Sw\n6pi8emhbNw75dKveqswt\n8dBqFRK93e9dHv7gsdl8\ncLbC7UgNJ2lirTRGfMO9",
}

var resultSet []string

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

func BenchmarkFormatPublicType(b *testing.B) {
	var localResultSet []string

	for n := 0; n < b.N; n++ {
		localResultSet = []string{}

		for i := 0; i < len(testSet); i++ {
			localResultSet = append(localResultSet, formatPublicType(testSet[i]))
		}
	}

	resultSet = localResultSet
}

func formatPublicType(name string) string {
	name = nonAlphanumericRegex.ReplaceAllString(name, "")

	if unicode.IsDigit(rune(name[0])) {
		name = "T" + name
	} else {
		name = strings.ToUpper(string(name[0])) + name[1:]
	}

	return name
}

var nonAlphanumericRegexRecursive = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func formatPublicTypeRecursive(name string) string {
	name = nonAlphanumericRegexRecursive.ReplaceAllString(name, "")

	if unicode.IsDigit(rune(name[0])) {
		name = "T" + name
	} else {
		name = strings.ToUpper(string(name[0])) + name[1:]
	}

	return name
}

func BenchmarkFormatPublicTypeRecursive(b *testing.B) {
	var localResultSet []string

	for n := 0; n < b.N; n++ {
		localResultSet = []string{}

		for i := 0; i < len(testSet); i++ {
			localResultSet = append(localResultSet, formatPublicTypeRecursive(testSet[i]))
		}
	}

	resultSet = localResultSet
}
