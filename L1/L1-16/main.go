package main

import "fmt"

func quickSort(slice []int) {
	if len(slice) < 2 {
		return
	} else {
		pivot := slice[0]
		left, right := 1, len(slice)-1

		for left <= right {
			switch {
			case slice[left] <= pivot:
				left++
				continue
			case slice[right] > pivot:
				right--
				continue
			default:
				slice[left], slice[right] = slice[right], slice[left]
				left++
				right--
			}
		}
		slice[0], slice[right] = slice[right], slice[0]
		quickSort(slice[:right])
		quickSort(slice[right+1:])
	}
}

func main() {
	slice := []int{3, 1, 5, 9, 5, 6, 7, 8, 9, 10}
	quickSort(slice)

	fmt.Println(slice)
}
