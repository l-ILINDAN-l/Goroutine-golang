package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Counter struct {
	atomic.Int64
}

func (c *Counter) Inc() {
	c.Add(1)
}

func main() {
	wg := new(sync.WaitGroup)
	counter := &Counter{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			counter.Inc()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			counter.Inc()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			counter.Inc()
		}
	}()
	wg.Wait()
	fmt.Println(counter.Load())
}
