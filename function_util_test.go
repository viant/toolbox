/*
 *
 *
 * Copyright 2012-2016 Viant.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 *  use this file except in compliance with the License. You may obtain a copy of
 *  the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 *  License for the specific language governing permissions and limitations under
 *  the License.
 *
 */
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
