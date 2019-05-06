package toolbox

import "sync"

//BatchLimiter represents a batch limiter
type BatchLimiter struct {
	queue chan uint8
	group *sync.WaitGroup
	Mutex *sync.RWMutex
}

//Acquire takes token form a channel, or wait if  no more elements in a a channel
func (r *BatchLimiter) Acquire() {
	<-r.queue
}

//Add adds element to wait group
func (r *BatchLimiter) Add(delta int) {
	r.group.Add(delta)
}

//Done flags wait group as done, returns back a token to a channel
func (r *BatchLimiter) Done() {
	r.group.Done()
	r.queue <- uint8(1)
}

//Wait wait on wait group
func (r *BatchLimiter) Wait() {
	r.group.Wait()
}

//NewBatchLimiter creates a new batch limiter with batch size and total number of elements
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
