package toolbox

import "sync"

type BatchLimiter struct {
	queue chan uint8
	group *sync.WaitGroup
	Mutex *sync.RWMutex
}

func (r *BatchLimiter) Acquire() {
	<-r.queue
}

func (r *BatchLimiter) Done() {
	r.group.Done()
	r.queue <- uint8(1)
}

func (r *BatchLimiter) Wait() {
	r.group.Wait()
}

func NewBatchLimiter(batchSize, total int) *BatchLimiter {
	var queue = make(chan uint8, batchSize)
	for i := 0; i < batchSize; i++ {
		queue <- uint8(1)
	}
	result := &BatchLimiter{
		queue: queue,
		group: &sync.WaitGroup{},
		Mutex: &sync.RWMutex{},
	}

	result.group.Add(total)
	return result
}
