package main

import (
	"fmt"
	"slices"
	"strings"
)

func anagrams(arrStrings []string) map[string][]string {
	mapping := make(map[string][]string)
	for _, str := range arrStrings {
		lowerStr := strings.ToLower(str)
		runes := []rune(lowerStr)
		slices.Sort(runes)
		mapping[string(runes)] = append(mapping[string(runes)], lowerStr)
	}
	result := make(map[string][]string)
	for _, slice := range mapping {
		if len(slice) > 1 {
			slices.Sort(slice)
			result[slice[0]] = slice
		}
	}
	return result
}

func main() {
	for key, slice := range anagrams([]string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}) {
		fmt.Printf("%q: %q\n", key, slice)
	}
}
