package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestSample struct {
	testString   string
	resultString string
	error        error
}

func TestUnpackingStr(t *testing.T) {
	tests := []TestSample{
		{
			testString:   "a4bc2d5e",
			resultString: "aaaabccddddde",
			error:        nil,
		},
		{
			testString:   "abcd",
			resultString: "abcd",
			error:        nil,
		},
		{
			testString:   "45",
			resultString: "",
			error:        ErrInvalidString,
		},
		{
			testString:   "",
			resultString: "",
			error:        nil,
		},
		{
			testString:   "qwe\\4\\5",
			resultString: "qwe45",
			error:        nil,
		},
		{
			testString:   "qwe\\45",
			resultString: "qwe44444",
			error:        nil,
		},
	}
	for _, test := range tests {
		testName := fmt.Sprintf("input: '%s'", test.testString)
		t.Run(testName, func(t *testing.T) {
			result, err := UnpackingString(test.testString)
			assert.Equal(t, test.resultString, result)
			assert.Equal(t, test.error, err)
		})
	}
}
