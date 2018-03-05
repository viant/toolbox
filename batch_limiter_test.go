package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
)

func TestBatchLimiter(t *testing.T) {
	var numbers = []int{1, 4, 6, 2, 5, 7, 5, 5, 7, 2, 3, 5, 7, 6}
	limiter := toolbox.NewBatchLimiter(4, len(numbers))
	var sum int32 = 0
	for _, n := range numbers {
		go func(n int32) {
			limiter.Acquire()
			defer limiter.Done()
			limiter.Mutex.Lock()
			defer limiter.Mutex.Unlock()
			//atomic.AddInt32(&sum, int32(n))
			sum = sum + int32(n)
		}(int32(n))

	}
	limiter.Wait()
	var expected int32 = 0
	for _, n := range numbers {
		expected += int32(n)
	}
	assert.Equal(t, expected, sum)
}
