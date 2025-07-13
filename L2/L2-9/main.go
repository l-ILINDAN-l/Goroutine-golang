package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ErrInvalidString is error about invalid string format
var ErrInvalidString = errors.New("invalid string format")

// UnpackingString is function unpacking string
func UnpackingString(s string) (string, error) {
	var result strings.Builder
	runes := []rune(s)

	for i := 0; i < len(runes); {
		char := runes[i]

		if unicode.IsDigit(char) {
			return "", ErrInvalidString
		}

		if char == '\\' {
			i++
			if i >= len(runes) {
				return "", ErrInvalidString
			}
			char = runes[i]
		}

		i++

		var numberBuilder strings.Builder
		for i < len(runes) && unicode.IsDigit(runes[i]) {
			numberBuilder.WriteRune(runes[i])
			i++
		}

		if numberBuilder.Len() > 0 {
			count, _ := strconv.Atoi(numberBuilder.String())
			result.WriteString(strings.Repeat(string(char), count))
		} else {
			result.WriteRune(char)
		}
	}

	return result.String(), nil
}

func main() {
	str, err := UnpackingString("a4bc4\\45\\5\\4")
	fmt.Println(str, err)
}
