package benchmark

import (
	"github.com/jamesits/goinvoke/utils"
	"testing"
)

/*
https://gosamples.dev/remove-non-alphanumeric/ offers an example of using regex `[^a-zA-Z0-9]+` to remove all the
non-alphanumeric characters from a string. I think it can be optimized by simply removing the `+` from the regex
eliminating the recursive lookup. But at that time I'm not sure how much of an improvement this small change
would bring.

Now I know.

goos: windows
goarch: amd64
pkg: github.com/jamesits/goinvoke/benchmark
cpu: AMD Ryzen 9 5900X 12-Core Processor
BenchmarkFormatPublicType
BenchmarkFormatPublicType-24                      179103              6502 ns/op
BenchmarkFormatPublicTypeRecursive
BenchmarkFormatPublicTypeRecursive-24             206896              5737 ns/op
*/

import (
	"regexp"
	"strings"
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

func BenchmarkFormatPublicType(b *testing.B) {
	var localResultSet []string

	for n := 0; n < b.N; n++ {
		localResultSet = []string{}

		for i := 0; i < len(testSet); i++ {
			localResultSet = append(localResultSet, utils.FormatPublicType(testSet[i]))
		}
	}

	resultSet = localResultSet
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
