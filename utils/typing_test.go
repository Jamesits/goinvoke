package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testSet = []string{
	"abcdefg",
	"Abcdefg",
	"te st ـا ą ٦测试文字あいう__*",
	"t\te��\r��s\nt",
	"1test",
}

var resultSet = []string{
	"Abcdefg",
	"Abcdefg",
	"Test",
	"Test",
	"T1test",
}

func TestFormatPublicType(t *testing.T) {
	for i := 0; i < len(testSet); i++ {
		assert.EqualValues(t, resultSet[i], FormatPublicType(testSet[i]))
	}
}
