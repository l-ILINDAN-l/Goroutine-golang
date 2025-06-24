package main

import (
	"fmt"
	"math"
	"sync"
)

func main() {
	testSlice := []int{2, 4, 6, 8, 10}
	var wg sync.WaitGroup

	for _, numSlice := range testSlice {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			fmt.Println(math.Pow(float64(num), 2))
		}(numSlice)
	}
	wg.Wait()
}
