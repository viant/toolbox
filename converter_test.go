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
	"testing"
	"github.com/viant/toolbox"
	"github.com/stretchr/testify/assert"
	"reflect"
)


func TestConverter(t *testing.T) {
	converter :=toolbox.NewColumnConverter("")
	{
		var value interface{};
		var test  = 123;
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, 123, *(value.(*int)))
	}
	{
		var value interface{};
		var test  = 123;
		err := converter.AssignConverted(&value, test)
		assert.Nil(t, err)
		assert.Equal(t, 123, value.(int))
	}

	{
		var value []byte;
		var test  = []byte("abc");
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", string(value))
	}

	{
		var value []byte;
		var test  = []byte("abc");
		err := converter.AssignConverted(&value, test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", string(value))
	}

	{
		var value string;
		err := converter.AssignConverted(&value, "abc")
		assert.Nil(t, err)
		assert.Equal(t, "abc", value)
	}

	{
		var value string;
		var test  = "abc";
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", value)
	}

	{

		var value *string;
		err := converter.AssignConverted(&value, "abc")
		assert.Nil(t, err)
		assert.Equal(t, "abc", *value)
	}

	{
		var value *string;
		var test  = "abc";
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", *value)
	}



	{
		var value string;
		err := converter.AssignConverted(&value, []byte("abc"))
		assert.Nil(t, err)
		assert.Equal(t, "abc", value)
	}

	{
		var value string;
		var test  = []byte("abc");
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", value)
	}

	{

		var value *string;
		err := converter.AssignConverted(&value, []byte("abc"))
		assert.Nil(t, err)
		assert.Equal(t, "abc", *value)
	}

	{
		var value *string;
		var test  = []byte("abc");
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", *value)
	}



	{
		var value int64;
		for _, item := range []interface{}{int(102),int64(102), float64(102), float32(102), "102"  } {
			err:=converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, int64(102), value)
		}
	}
	{
		var value *int64;
		for _, item := range []interface{}{int(102),int64(102), float64(102), float32(102), "102"  } {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, int64(102), *value)
		}
	}

	{
		var value *float64;
		for _, item := range []interface{}{int(102),int64(102), float64(102), float32(102), "102"  } {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, float64(102), *value)
		}
	}

}


func TestDiscoverCollectionValueType(t *testing.T) {
	{
		var input= []string{"3.2", "1.2"}
		var output, kind = toolbox.DiscoverCollectionValuesAndKind(input)
		assert.Equal(t, reflect.Float64, kind)
		assert.Equal(t, 1.2, output[1])
	}
	{
		var input = []string{"3.2", "abc"}
		var output, kind = toolbox.DiscoverCollectionValuesAndKind(input)
		assert.Equal(t, reflect.String, kind)
		assert.Equal(t, "abc", output[1])
		assert.Equal(t, "3.2", output[0])
	}
}