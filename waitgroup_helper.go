package toolbox

import (
	"sync"
	"sync/atomic"
	"time"
)

// WaitGroup that waits with a timeout
// Returns true if timeout exceeded and false if there was no timeout
func WaitTimeout(wg *sync.WaitGroup, duration time.Duration) bool {
	done := make(chan bool, 1)
	closed := int32(0)
	defer func() {
		if atomic.CompareAndSwapInt32(&closed, 0, 1) {
			close(done)
		}
	}()
	go func() {
		wg.Wait()
		if atomic.LoadInt32(&closed) == 0 {
			done <- true
		}
	}()

	select {
	case <-done: //Wait till the task is complete and channel get unblocked
		return false //No durationToken. Normal execution of task completion
	case <-time.After(duration): //Wait till durationToken to elapse
		//TODO: time.After() creates a timer that does not get GC until timer durationToken gets elapsed. Need to use AfterFunc
		return true //Timed out
	}
}
