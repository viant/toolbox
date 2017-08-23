package bridge_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
)

func TestNewBytesBufferPool(t *testing.T) {
	pool := toolbox.NewBytesBufferPool(2, 1024)
	buf := pool.Get() //creates a new buffer pool is empty.
	assert.NotNil(t, buf)

	buf[1] = 0x2
	pool.Put(buf)
	pool.Put([]byte{0x1})
	pool.Put(buf) //get discarded
	{
		poolBuf := pool.Get()
		assert.Equal(t, uint8(0x2), poolBuf[1])
		assert.Equal(t, 1024, len(poolBuf))
	}
	{
		poolBuf := pool.Get()
		assert.Equal(t, 1, len(poolBuf))
	}
}
