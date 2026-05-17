package task

import (
	"fmt"
	"log"
)

type Task interface {
	Execute() error
}

type Queue struct {
	taskChan chan Task
}

func NewQueue(size int) *Queue {
	return &Queue{
		taskChan: make(chan Task, size),
	}
}

func (q *Queue) Enqueue(task Task) error {
	select {
	case q.taskChan <- task:
		return nil
	default:
		return fmt.Errorf("queue full")
	}
}

func (q *Queue) StartWorkers(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go q.worker()
	}
}

func (q *Queue) worker() {
	for task := range q.taskChan {
		if err := task.Execute(); err != nil {
			log.Printf("Task execution failed: %v", err)
		}
	}
}

func (q *Queue) Close() {
	close(q.taskChan)
}
