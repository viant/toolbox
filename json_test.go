package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"strings"
	"testing"
)

func Test_IsCompleteJSON(t *testing.T) {

	{
		input := `{"a":1, "b":2}`
		assert.True(t, toolbox.IsCompleteJSON(input))
	}
	{
		input := `{"a":1, "b":2}
{"a2":2, "b3":21}
{"a3":3, "b4:22
`
		assert.False(t, toolbox.IsCompleteJSON(input))
	}

}

func Test_IsNewDelimitedJSON(t *testing.T) {

	{
		input := `{"a":1, "b":2}`
		assert.False(t, toolbox.IsNewLineDelimitedJSON(input))
	}
	{
		input := `{"a":1, "b":2}
{"a2":2, "b3":21}
{"a3":3, "b4:22}
`
		assert.True(t, toolbox.IsNewLineDelimitedJSON(input))
	}
	{
		input := `{"a":1, "b":2}
{"a2":2, "b3":21
{"a3":3, "b4:22}
`
		assert.False(t, toolbox.IsNewLineDelimitedJSON(input))
	}

}

func Test_JSONToMap(t *testing.T) {
	{
		input := `{"a":1, "b":2}`
		aMAp, err := toolbox.JSONToMap(input)
		assert.Nil(t, err)
		assert.True(t, len(aMAp) > 0)
	}
	{
		input := `{"a":1, "b":2}`
		aMAp, err := toolbox.JSONToMap([]byte(input))
		assert.Nil(t, err)
		assert.True(t, len(aMAp) > 0)
	}
	{
		input := `{"a":1, "b":2}`
		aMAp, err := toolbox.JSONToMap(strings.NewReader(input))
		assert.Nil(t, err)
		assert.True(t, len(aMAp) > 0)
	}
	{
		//error case
		_, err := toolbox.JSONToMap(1)
		assert.NotNil(t, err)
	}
	{
		//error case
		input := `{"a":1, "b":2`
		_, err := toolbox.JSONToMap(input)
		assert.NotNil(t, err)
	}

}
