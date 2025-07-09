package main

import "fmt"

func interSection(slices ...[]int) []int {
	hashMap := make(map[int]int)
	for _, slice := range slices {
		nonDuplicate := make(map[int]struct{})
		for _, value := range slice {
			nonDuplicate[value] = struct{}{}
		}
		for key := range nonDuplicate {
			hashMap[key] += 1
		}
	}
	result := make([]int, 0)
	for key, value := range hashMap {
		if value == len(slices) {
			result = append(result, key)
		}

	}
	return result
}

func main() {
	slice1 := []int{1, 2, 3}
	slice2 := []int{2, 3, 4}
	fmt.Println(interSection(slice1, slice2))
}
