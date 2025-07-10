package main

import "fmt"

func binSearch(slice []int, value int) int {
	leftIndex, rightIndex := 0, len(slice)-1
	for leftIndex <= rightIndex {
		mid := (leftIndex + rightIndex) / 2
		switch {
		case slice[mid] == value:
			return mid
		case slice[mid] < value:
			leftIndex = mid + 1
		case slice[mid] > value:
			rightIndex = mid - 1
		}
	}
	return -1
}

func main() {
	slice := []int{1, 2, 3, 4, 5, 7, 8, 9, 10}
	fmt.Println(binSearch(slice, 5))
	fmt.Println(binSearch(slice, 6))
}
