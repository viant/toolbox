package toolbox_test

import (
	"testing"
	"github.com/viant/toolbox"
	"github.com/stretchr/testify/assert"
)

func TestSliceIterator(t *testing.T) {
	slice := []string{"a", "z", "c"}
	iterator := toolbox.NewSliceIterator(slice)
	value := ""

	assert.True(t, iterator.HasNext())
	iterator.Next(&value)
	assert.Equal(t, "a", value)

	assert.True(t, iterator.HasNext())
	iterator.Next(&value)
	assert.Equal(t, "z", value)

	assert.True(t, iterator.HasNext())
	iterator.Next(&value)
	assert.Equal(t, "c", value)

}

