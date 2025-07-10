package main

import (
	"fmt"
	"time"
)

func mySleep(d time.Duration) {
	<-time.After(d)
}

func main() {
	fmt.Println("mySleep на 2 c")
	start1 := time.Now()
	mySleep(2 * time.Second)
	fmt.Printf("Прошло времени: %v\n\n", time.Since(start1))

	fmt.Println("mySleep на 3 с")
	start2 := time.Now()
	mySleep(3 * time.Second)
	fmt.Printf("Прошло времени: %v\n", time.Since(start2))
}
