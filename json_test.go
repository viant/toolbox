package toolbox_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
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


