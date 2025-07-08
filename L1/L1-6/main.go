package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func goroutineByCondition(numSteps int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Start goroutine exit by condition...")
	for step := 1; step <= numSteps; step++ {
		fmt.Printf("goroutineByCondition step: %d\n", step)
	}
	fmt.Println("end goroutine exit by condition...")
}

func goroutineByNotificationChannel(channel <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Start goroutine exit by notification channel...")
	for {
		select {
		case <-channel:
			fmt.Println("goroutineByNotificationChannel exit by notification channel...")
			return
		default:
			fmt.Println("goroutineByNotificationChannel continue")
			// Process something
			time.Sleep(1 * time.Second)
		}
	}
}

func goroutineByContext(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Start goroutine exit by context...")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("goroutineByContext exit by context...")
			return
		default:
			fmt.Println("goroutineByContext continue")
			// Process something
			time.Sleep(1 * time.Second)
		}
	}
}

func goroutineByGoExit(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Start goroutine exit by goExit()...")
	for i := 0; i < 100; i++ {
		if i == 10 {
			fmt.Println("goroutineByGoExit exit by GoExit...")
			runtime.Goexit()
		}
		fmt.Println("goroutineByGoExit continue")
		// Process something
		time.Sleep(1 * time.Second)
	}
	fmt.Println("goroutineByGoExit exit by condition...")
}

func goroutineByCloseChannel(channel <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Start goroutine exit by closing channel...")
	for i := range channel {
		fmt.Println("goroutineByCloseChannel: ", i)
		// Process something
		time.Sleep(1 * time.Second)
	}
	fmt.Println("goroutineByCloseChannel exit by close channel...")
}

func goroutineByPanic(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		if err := recover(); err != nil {
			fmt.Println("goroutineByPanic panic:", err)
		}
	}()
	for i := 0; i < 100; i++ {
		if i == 10 {
			panic("goroutineByPanic exit by panic...")
		}
		fmt.Println("goroutineByPanic continue")
		// Process something
		time.Sleep(1 * time.Second)
	}
	fmt.Println("goroutineByPanic exit by condition...")
}

func goroutineByAtomicFlag(flag *atomic.Bool, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Start goroutine exit by atomic flag...")
	for {
		if flag.Load() {
			fmt.Println("goroutineByAtomicFlag exit by atomic flag...")
			return
		}
		fmt.Println("goroutineByAtomicFlag continue")
		// Process something
		time.Sleep(1 * time.Second)
	}
}

func main() {
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go goroutineByCondition(10, wg)

	wg.Add(1)
	notifChann := make(chan struct{})
	go goroutineByNotificationChannel(notifChann, wg)

	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go goroutineByContext(ctx, wg)

	wg.Add(1)
	go goroutineByGoExit(wg)

	wg.Add(1)
	workChann := make(chan int, 10)
	for i := 0; i < 10; i++ {
		workChann <- i
	}
	go goroutineByCloseChannel(workChann, wg)

	wg.Add(1)
	go goroutineByPanic(wg)

	wg.Add(1)
	flag := new(atomic.Bool)
	go goroutineByAtomicFlag(flag, wg)

	close(notifChann)
	cancel()
	close(workChann)
	flag.Store(true)

	wg.Wait()
}
