package main

import (
	"fmt"
	"sync"
)

type Worker struct {
	index    int
	workChan <-chan string
	quit     chan struct{}
	wg       *sync.WaitGroup
}

func NewWorker(index int, workChan <-chan string, quit chan struct{}, wg *sync.WaitGroup) *Worker {
	return &Worker{
		index:    index,
		workChan: workChan,
		quit:     quit,
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
		case msg := <-w.workChan:
			w.Work(msg)
		case <-w.quit:
			return
		}
	}
}

func (w *Worker) Stop() {
	close(w.quit)
}

type WorkerPool struct {
	workerPool map[int]*Worker
	maxIndex   int
	workChan   <-chan string
	wg         *sync.WaitGroup
}

func NewWorkerPool(numWorker int, workChan <-chan string) *WorkerPool {
	workerPoolMap := make(map[int]*Worker)
	wg := new(sync.WaitGroup)
	for index := 0; index < numWorker; index++ {
		workerPoolMap[index] = NewWorker(index, workChan, make(chan struct{}), wg)
	}
	return &WorkerPool{
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

func (wp *WorkerPool) Stop() {
	for i := 0; i < wp.maxIndex; i++ {
		wp.workerPool[i].Stop()
	}
	wp.wg.Wait()
}

func main() {
	chanWork := make(chan string)
	fmt.Print("Enter number workers: ")
	var numWorker int
	if _, err := fmt.Scan(&numWorker); err != nil {
		fmt.Println("Error: not valid value of number workers ", err)
	}
	fmt.Println("Enter 'quit' to stop application, another message is string for processing")

	workerPool := NewWorkerPool(numWorker, chanWork)
	workerPool.Start()
	for {
		var msgCLI string
		if _, err := fmt.Scanln(&msgCLI); err != nil {
			fmt.Println("Error: not valid value of command ", err)
			return
		}
		if msgCLI == "quit" {
			workerPool.Stop()
			break
		} else {
			chanWork <- msgCLI
		}
	}
}
