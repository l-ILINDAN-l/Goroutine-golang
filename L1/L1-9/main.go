package main

import (
	"fmt"
	"sync"
)

func Multiply2(chanNums <-chan int, chanOutput chan<- int, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		close(chanOutput)
	}()
	for num := range chanNums {
		chanOutput <- num * 2
	}
}

func PrintFromChannel(printChan <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for num := range printChan {
		fmt.Println(num)
	}
}

func main() {
	chanNums := make(chan int)
	chanMultiply2 := make(chan int)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go Multiply2(chanNums, chanMultiply2, wg)
	go PrintFromChannel(chanMultiply2, wg)

	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, num := range arr {
		chanNums <- num
	}
	close(chanNums)

	wg.Wait()
}
