package toolbox_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func Test_IsCompleteJSON(t *testing.T) {
	/*
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
	   	}*/
	{
		input := `{"name":"abc"},{"id":"10}"`
		assert.True(t, toolbox.IsCompleteJSON(input))
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

func Test_AsJSONText(t *testing.T) {
	{
		var soure = map[string]interface{}{
			"k": 1,
		}
		text, err := toolbox.AsJSONText(soure)
		assert.Nil(t, err)
		assert.EqualValues(t, "{\"k\":1}\n", text)
	}
	{
		type source struct {
			K int
		}
		text, err := toolbox.AsJSONText(&source{K: 1})
		assert.Nil(t, err)
		assert.EqualValues(t, "{\"K\":1}\n", text)
	}

	{

		text, err := toolbox.AsJSONText([]int{1, 3})
		assert.Nil(t, err)
		assert.EqualValues(t, "[1,3]\n", text)
	}

	{

		_, err := toolbox.AsJSONText(1)
		assert.NotNil(t, err)
	}

}

func Test_JSONToInterface(t *testing.T) {
	{
		input := `{"a":1, "b":2}`
		output, err := toolbox.JSONToInterface(input)
		if assert.Nil(t, err) {
			assert.NotNil(t, output)
			assert.True(t, toolbox.IsMap(output))
			aMap := toolbox.AsMap(output)
			assert.EqualValues(t, 1, aMap["a"])
			assert.EqualValues(t, 2, aMap["b"])
		}
	}
	{
		input := `[1,2]`
		output, err := toolbox.JSONToInterface(input)
		if assert.Nil(t, err) {
			assert.NotNil(t, output)
			assert.True(t, toolbox.IsSlice(output))
			aSlice := toolbox.AsSlice(output)
			assert.EqualValues(t, []interface{}{1.0, 2.0}, aSlice)
		}
	}
}
