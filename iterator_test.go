package toolbox_test

import (
	"testing"
	"github.com/viant/toolbox"
	"github.com/stretchr/testify/assert"
)

func TestSliceIterator(t *testing.T) {

	{
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
	{
		slice := []interface{}{"a", "z", "c"}
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

	{
		slice := []int{3, 2, 1}
		iterator := toolbox.NewSliceIterator(slice)
		value := 0
		assert.True(t, iterator.HasNext())
		iterator.Next(&value)
		assert.Equal(t, 3, value)

		assert.True(t, iterator.HasNext())
		iterator.Next(&value)
		assert.Equal(t, 2, value)

		assert.True(t, iterator.HasNext())
		iterator.Next(&value)
		assert.Equal(t, 1, value)
	}

}

