package main

import (
	"fmt"
	"strings"
)

var justString string

func createHugeString(size int) string {
	return strings.Repeat("*", size)
}

func someFunc() {
	v := createHugeString(1 << 10)
	justString = strings.Clone(v[:100])
}

func main() {
	someFunc()
	fmt.Println(justString)
}
