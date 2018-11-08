package toolbox

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitTimeout_TimeoutTriggered(t *testing.T) {
	var wg sync.WaitGroup
	//Sleep for 3 seconds
	go sleep(&wg)

	//task will sleep for 3 seconds but timeout is set only for 1 second
	isTimeOut := WaitTimeout(&wg, time.Second*1)
	assert.Equal(t, true, isTimeOut)
}

func TestWaitTimeout_NoTimeout(t *testing.T) {
	var wg sync.WaitGroup

	//Sleep for 3 seconds
	go sleep(&wg)

	//task will sleep for 3 seconds but timeout is set only for 5 second
	isTimeOut := WaitTimeout(&wg, time.Second*5)
	assert.Equal(t, false, isTimeOut)
}

//Method that sleeps for 3 seconds
func sleep(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	time.Sleep(time.Second * 3)
}
