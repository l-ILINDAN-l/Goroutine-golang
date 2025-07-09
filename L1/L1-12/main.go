package main

import "fmt"

func fromSliceToSet(slice []string) []string {
	hashMap := make(map[string]struct{})
	for _, v := range slice {
		hashMap[v] = struct{}{}
	}
	resultSet := make([]string, 0, len(hashMap))
	for k := range hashMap {
		resultSet = append(resultSet, k)
	}
	return resultSet
}

func main() {
	fmt.Print(fromSliceToSet([]string{"cat", "cat", "dog", "cat", "tree"}))
}
