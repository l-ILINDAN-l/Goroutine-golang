package main

import (
	"fmt"
	"sync"
	"time"
)

type Result string

type SafeMap struct {
	sync.RWMutex
	mapRes map[string]Result
}

var cache = SafeMap{mapRes: make(map[string]Result)}

func ProcessData(str string) Result {
	cache.RLock()
	res, ok := cache.mapRes[str]
	cache.RUnlock()
	if ok {
		return res
	}

	cache.Lock()
	defer cache.Unlock()

	res, ok = cache.mapRes[str]
	if ok {
		return res
	}

	res = Result(str + "_result")
	time.Sleep(10 * time.Second)
	cache.mapRes[str] = res
	return cache.mapRes[str]
}

func main() {
	wg := sync.WaitGroup{}

	wg.Add(2)
	go func() {
		strings := []string{"1111", "2222", "3333", "4444", "5555"}
		for _, str := range strings {
			fmt.Println(ProcessData(str))
		}
		wg.Done()
	}()

	go func() {
		strings := []string{"3333", "4444", "5555", "6666", "7777"}
		for _, str := range strings {
			fmt.Println(ProcessData(str))
		}
		wg.Done()
	}()

	wg.Wait()
}
