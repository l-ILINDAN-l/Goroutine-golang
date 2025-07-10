package main

import (
	"fmt"
	"strings"
)

func reverse(s []rune) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func reverseWords(s string) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	reverse(runes)
	wordStart := 0
	//for wordStart < len(runes) {
	//	for end := wordStart; end <= len(runes); end++ {
	//		if end == len(runes) || runes[end] == ' ' {
	//			reverse(runes[wordStart:end])
	//			end++
	//			wordStart = end
	//			break
	//		}
	//	}
	//}
	for i := 0; i <= len(runes); i++ {
		if i == len(runes) || runes[i] == ' ' {
			reverse(runes[wordStart:i])
			wordStart = i + 1
		}
	}
	return string(runes)
}

func main() {
	str := "snow dog sun"
	fmt.Println(reverseWords(str))
}
