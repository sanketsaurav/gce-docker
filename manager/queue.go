package manager

import (
	"sync"
	"time"
)

type JobID string
type Job func() error

type Worker struct {
	jobs map[JobID]Job
	sync.Mutex
}

func NewWorker() *Worker {
	return &Worker{
		jobs: make(map[JobID]Job, 0),
	}
}

func (w *Worker) Add(id JobID, j Job, delay time.Duration) {
	w.Lock()
	defer w.Unlock()

	w.jobs[id] = j
	go w.do(id, delay)
}

func (w *Worker) do(id JobID, delay time.Duration) {
	<-time.After(delay)
	defer w.Delete(id)

	if j, ok := w.jobs[id]; ok {
		j()
	}
}

func (w *Worker) Delete(id JobID) bool {
	w.Lock()
	defer w.Unlock()

	_, ok := w.jobs[id]
	delete(w.jobs, id)

	return ok
}
