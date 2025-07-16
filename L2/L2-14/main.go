package main

import (
	"fmt"
	"sync"
	"time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {
	resChan := make(chan interface{})
	onceCloseChan := sync.Once{}

	done := make(chan struct{})
	for _, channel := range channels {
		go func(channel <-chan interface{}) {
			for {
				select {
				case <-channel:
					onceCloseChan.Do(func() {
						close(resChan)
						close(done)
					})
					return
				case <-done:
					return
				}
			}
		}(channel)
	}
	return resChan
}

func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(1*time.Second),
		sig(1*time.Second),
		sig(1*time.Second),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))
}
