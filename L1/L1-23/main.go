package main

import "fmt"

func deleteByIndex(slice []int, index int) []int {
	copy(slice[index:], slice[index+1:])
	slice[len(slice)-1] = 0 // nil для среза указателей и "" для строк
	return slice[:len(slice)-1]
}

func main() {
	slice := []int{1, 2, 3}
	slice = deleteByIndex(slice, 1)
	fmt.Println(slice)
}
