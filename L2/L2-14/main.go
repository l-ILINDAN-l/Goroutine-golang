package main

import (
	"fmt"
	"sync"
	"time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {
	resChan := make(chan interface{})
	onceCloseChan := sync.Once{}

	//ctx, cancel := context.WithCancel(context.Background())
	for _, channel := range channels {
		go func(channel <-chan interface{}) {
			for {
				select {
				case <-channel:
					onceCloseChan.Do(func() {
						close(resChan)
						//cancel()
					})
					return
					// Хотел добавить сюда закрытие горутин других каналов, если хотя бы одна закрылась,
					// но линтер ругался при использовании контекста, что cancel() может в теоретическом случае не вызваться.
					// Если подскажите, как нерекурсивно напить "идеально", буду благодарен
					//case <-ctx.Done():
					//	return
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
