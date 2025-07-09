package main

import (
	"fmt"
)

func main() {
	// Solve O(n)
	tempatures := []float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5}
	groups := make(map[int][]float64)
	for _, num := range tempatures {
		groups[int(num/10)*10] = append(groups[int(num/10)*10], num)
	}
	for key, val := range groups {
		fmt.Println(key, val)
	}
}
