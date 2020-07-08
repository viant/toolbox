package toolbox

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)


func TestWaitTimeout_TimeoutTriggered(t *testing.T) {
	var wg sync.WaitGroup
	go sleep(&wg, time.Second)
	time.Sleep(100 * time.Millisecond)
	isTimeOut := WaitTimeout(&wg, 100 * time.Millisecond)
	assert.Equal(t, true, isTimeOut)
}

func TestWaitTimeout_NoTimeout(t *testing.T) {
	var wg sync.WaitGroup
	go sleep(&wg, 100 * time.Millisecond)
	//task will sleep for 3 seconds but timeout is set only for 5 second
	isTimeOut := WaitTimeout(&wg, time.Second)
	assert.Equal(t, false, isTimeOut)
}

//Method that sleeps for 3 seconds
func sleep(wg *sync.WaitGroup, duration time.Duration) {
	wg.Add(1)
	time.Sleep(duration)
	wg.Done()
}
