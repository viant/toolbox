package toolbox_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestCallFunction(t *testing.T) {

	{
		var myFunction = func(arg1 string, arg2 int) string {
			return fmt.Sprintf("%v%v", arg2, arg1)
		}

		functionParameters, err := toolbox.BuildFunctionParameters(myFunction, []string{"arg1", "arg2"}, map[string]interface{}{
			"arg1": "abc",
			"arg2": 100,
		})

		assert.Nil(t, err)
		result := toolbox.CallFunction(myFunction, functionParameters...)

		assert.Equal(t, "100abc", result[0])
	}
	{

		var myFunction = func(arg1 string, arg2 ...int) string {
			return fmt.Sprintf("%v%v", arg2, arg1)
		}

		functionParameters, err := toolbox.BuildFunctionParameters(myFunction, []string{"arg1", "arg2"}, map[string]interface{}{
			"arg1": "abc",
			"arg2": []interface{}{100},
		})

		assert.Nil(t, err)
		result := toolbox.CallFunction(myFunction, functionParameters...)

		assert.Equal(t, "[100]abc", result[0])
	}
	{

		var myFunction = func(arg1 string, arg2 ...int) string {
			return fmt.Sprintf("%v%v", arg2, arg1)
		}

		_, err := toolbox.BuildFunctionParameters(myFunction, []string{"arg1", "arg2"}, map[string]interface{}{
			"arg1": "abc",
			"arg2": 100,
		})

		assert.NotNil(t, err)
	}
}

func Test_GetFunction(t *testing.T) {
	var astruct = &AStruct{"ABC"}
	var function, err = toolbox.GetFunction(astruct, "Message")
	assert.Nil(t, err)
	assert.NotNil(t, function)
	parameters, err := toolbox.AsCompatibleFunctionParameters(function, []interface{}{"aaa"})
	assert.Nil(t, err)
	result := toolbox.CallFunction(function, parameters...)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "ABC.aaa", result[0])

}

type AStruct struct {
	A string
}

func (s *AStruct) Message(a string) (string, error) {
	return fmt.Sprintf("%v.%v", s.A, a), nil
}
