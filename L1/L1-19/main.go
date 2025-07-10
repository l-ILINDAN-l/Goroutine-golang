package main

import "fmt"

func Reverse(s string) string {
	runeSlice := []rune(s)
	for i, j := 0, len(runeSlice)-1; i < j; i, j = i+1, j-1 {
		runeSlice[i], runeSlice[j] = runeSlice[j], runeSlice[i]
	}
	return string(runeSlice)
}

func main() {
	str := "абырвалг😃"
	fmt.Println(Reverse(str))
}
