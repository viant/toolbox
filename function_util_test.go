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
