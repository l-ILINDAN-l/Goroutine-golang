package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
)

type Worker struct {
	ctx      context.Context
	index    int
	workChan <-chan string
	wg       *sync.WaitGroup
}

func NewWorker(ctx context.Context, index int, workChan <-chan string, wg *sync.WaitGroup) *Worker {
	return &Worker{
		ctx:      ctx,
		index:    index,
		workChan: workChan,
		wg:       wg,
	}
}
func (w *Worker) Work(str string) {
	fmt.Println(str)
}

func (w *Worker) Start() {
	defer w.wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case msg, ok := <-w.workChan:
			if !ok {
				return
			}
			w.Work(msg)

		}
	}
}

type WorkerPool struct {
	ctx        context.Context
	workerPool map[int]*Worker
	maxIndex   int
	workChan   <-chan string
	wg         *sync.WaitGroup
}

func NewWorkerPool(ctx context.Context, numWorker int, workChan <-chan string) *WorkerPool {
	workerPoolMap := make(map[int]*Worker)
	wg := new(sync.WaitGroup)
	for index := 0; index < numWorker; index++ {
		workerPoolMap[index] = NewWorker(ctx, index, workChan, wg)
	}
	return &WorkerPool{
		ctx:        ctx,
		workerPool: workerPoolMap,
		maxIndex:   numWorker,
		workChan:   workChan,
		wg:         wg,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.maxIndex; i++ {
		wp.wg.Add(1)
		go wp.workerPool[i].Start()
	}
}
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

func main() {
	fmt.Print("Enter number workers: ")
	var numWorker int
	if _, err := fmt.Scan(&numWorker); err != nil {
		fmt.Println("Error: not valid value of number workers ", err)
	}
	fmt.Println("Enter 'quit' to stop application, another message is string for processing")

	chanWork := make(chan string)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	defer cancel()

	workerPool := NewWorkerPool(ctx, numWorker, chanWork)
	workerPool.Start()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msgCLI string
				if _, err := fmt.Scanln(&msgCLI); err != nil {
					cancel()
					return
				}
				if msgCLI == "quit" {
					cancel()
					return
				}

				chanWork <- msgCLI
			}

		}
	}()
	<-ctx.Done()
	close(chanWork)
	workerPool.Wait()
}
