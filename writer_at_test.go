package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
)

func TestWriterAt_WriteAt(t *testing.T) {
	writer := toolbox.NewByteWriterAt()
	writer.WriteAt([]byte{0x2}, 1)
	writer.WriteAt([]byte{0x1}, 0)
	writer.WriteAt([]byte{0x3}, 2)
	assert.Equal(t, []byte{0x1, 0x02, 0x3}, writer.Buffer)
}
