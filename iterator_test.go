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

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestSliceIterator(t *testing.T) {

	{
		slice := []string{"a", "r", "c"}
		iterator := toolbox.NewSliceIterator(slice)
		var values = make([]interface{}, 1)
		value := values[0]

		assert.True(t, iterator.HasNext())
		iterator.Next(&value)
		assert.Equal(t, "a", value)

		assert.True(t, iterator.HasNext())
		iterator.Next(&value)
		assert.Equal(t, "r", value)

		assert.True(t, iterator.HasNext())
		iterator.Next(&value)

		assert.Equal(t, "c", value)

	}
	{
		slice := []string{"a", "r", "c"}
		iterator := toolbox.NewSliceIterator(slice)
		value := ""

		assert.True(t, iterator.HasNext())
		iterator.Next(&value)
		assert.Equal(t, "a", value)

		assert.True(t, iterator.HasNext())
		iterator.Next(&value)
		assert.Equal(t, "r", value)

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

		var values = make([]interface{}, 1)
		assert.True(t, iterator.HasNext())
		iterator.Next(&values[0])
		assert.Equal(t, "c", values[0])

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
