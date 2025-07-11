package main

import (
	"fmt"
	"strings"
)

func isUniqChar(s string) bool {
	mapping := make(map[rune]struct{})
	s = strings.ToLower(s)
	for _, c := range s {
		if _, ok := mapping[c]; ok {
			return false
		} else {
			mapping[c] = struct{}{}
		}
	}
	return true
}

func main() {
	fmt.Println(isUniqChar("abcd"))
	fmt.Println(isUniqChar("abCdefAaf"))
	fmt.Println(isUniqChar("abCdefAaf"))
}
