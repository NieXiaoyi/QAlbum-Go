package task

import (
	"context"
	"log"
	"time"
)

type Worker struct {
	queues      []*Queue
	ctx         context.Context
	cancel      context.CancelFunc
	workerCount int
}

func NewWorker(ctx context.Context, workerCount int, queues ...*Queue) *Worker {
	workerCtx, cancel := context.WithCancel(ctx)
	return &Worker{
		queues:      queues,
		ctx:         workerCtx,
		cancel:      cancel,
		workerCount: workerCount,
	}
}

func (w *Worker) Start() {
	for _, queue := range w.queues {
		for i := 0; i < w.workerCount; i++ {
			go w.worker(queue, i)
		}
	}
}

func (w *Worker) worker(queue *Queue, workerID int) {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		select {
		case task := <-queue.taskChan:
			if err := task.Execute(); err != nil {
				log.Printf("[WORKER-%d] Task execution failed: %v", workerID, err)
			} else {
				log.Printf("[WORKER-%d] Task executed successfully", workerID)
			}
		case <-w.ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (w *Worker) Stop() {
	w.cancel()
}
