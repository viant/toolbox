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
	"reflect"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestConverter(t *testing.T) {
	converter := toolbox.NewColumnConverter(toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z"))

	{
		target := make([]interface{}, 1)
		err := converter.AssignConverted(&target[0], nil)
		assert.Nil(t, err)
		assert.Nil(t, target[0])
	}

	{
		err := converter.AssignConverted(nil, nil)
		assert.NotNil(t, err)
	}

	{
		var value interface{}
		var test = 123
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, 123, *(value.(*int)))
	}
	{
		var value interface{}
		var test = 123
		err := converter.AssignConverted(&value, test)
		assert.Nil(t, err)
		assert.Equal(t, 123, value.(int))
	}

	{
		//Byte types
		{
			var value []byte
			var test = []byte("abc")
			err := converter.AssignConverted(&value, &test)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(value))
		}
		{
			var value []byte
			var test = "abc"
			err := converter.AssignConverted(&value, &test)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(value))
		}

		{
			var value []byte
			var test = []byte("abc")
			err := converter.AssignConverted(&value, test)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(value))
		}

		{
			var value []byte
			var test = "abc"
			err := converter.AssignConverted(&value, test)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(value))
		}

	}

	{
		//Byte types
		{
			var value *[]byte
			var test = []byte("abc")
			err := converter.AssignConverted(&value, &test)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(*value))
		}
		{
			var value *[]byte
			var test = "abc"
			err := converter.AssignConverted(&value, &test)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(*value))
		}

		{
			var value *[]byte
			var test = []byte("abc")
			err := converter.AssignConverted(&value, test)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(*value))
		}

		{
			var value *[]byte
			var test = "abc"
			err := converter.AssignConverted(&value, test)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(*value))
		}

	}

	{
		var value string
		err := converter.AssignConverted(&value, "abc")
		assert.Nil(t, err)
		assert.Equal(t, "abc", value)
	}

	{
		var value string
		var test = "abc"
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", value)
	}

	{
		var value string
		var test = 12
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "12", value)
	}

	{
		var value *string
		var test = 12
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "12", *value)
	}

	{

		var value *string
		err := converter.AssignConverted(&value, "abc")
		assert.Nil(t, err)
		assert.Equal(t, "abc", *value)
	}

	{
		var value *string
		var test = "abc"
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", *value)
	}

	{
		var value string
		err := converter.AssignConverted(&value, []byte("abc"))
		assert.Nil(t, err)
		assert.Equal(t, "abc", value)
	}

	{
		var value string
		var test = []byte("abc")
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", value)
	}

	{

		var value *string
		err := converter.AssignConverted(&value, []byte("abc"))
		assert.Nil(t, err)
		assert.Equal(t, "abc", *value)
	}

	{
		var value *string
		var test = []byte("abc")
		err := converter.AssignConverted(&value, &test)
		assert.Nil(t, err)
		assert.Equal(t, "abc", *value)
	}

	{
		var value float64
		for _, item := range []interface{}{int(102), int64(102), float64(102), float32(102), "102"} {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, float64(102), value)
		}
	}

	{
		var value int64
		for _, item := range []interface{}{int(102), int64(102), float64(102), float32(102), "102"} {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, int64(102), value)
		}
	}
	{
		var value *int64
		for _, item := range []interface{}{int(102), int64(102), float64(102), float32(102), "102"} {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, int64(102), *value)
		}
	}

	{
		var value uint64
		for _, item := range []interface{}{int(102), int64(102), float64(102), float32(102), "102"} {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, uint64(102), value)
		}
	}
	{
		var value *uint64
		for _, item := range []interface{}{int(102), int64(102), float64(102), float32(102), "102"} {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, uint64(102), *value)
		}
	}

	{
		var value *float64
		for _, item := range []interface{}{int(102), int64(102), float64(102), float32(102), "102"} {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.Equal(t, float64(102), *value)
		}
	}

	{
		var value *bool
		sTrue := "true"
		vTrue := true

		for _, item := range []interface{}{1, true, "true", &sTrue, &vTrue} {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.True(t, *value)
		}
		err := converter.AssignConverted(&value, "abc")
		assert.NotNil(t, err)

	}
	{
		var value bool
		sTrue := "true"
		vTrue := true
		for _, item := range []interface{}{1, true, "true", &sTrue, &vTrue} {
			err := converter.AssignConverted(&value, item)
			assert.Nil(t, err)
			assert.True(t, value)
		}
		err := converter.AssignConverted(&value, "abc")
		assert.NotNil(t, err)

	}
	{

		var value *time.Time
		date := "2016-02-22 12:32:01 UTC"
		{
			err := converter.AssignConverted(&value, date)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			err := converter.AssignConverted(&value, &date)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			err := converter.AssignConverted(&value, "1456144321")
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			unix := "1456144321.0"
			err := converter.AssignConverted(&value, unix)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			unix := 1456144321.0
			err := converter.AssignConverted(&value, unix)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			unix := 1456144321.0
			err := converter.AssignConverted(&value, &unix)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			date := "2016/02/22 12:32:01 UTC"
			err := converter.AssignConverted(&value, date)
			assert.NotNil(t, err, "invalid date format")
		}

	}

	{

		var value time.Time
		date := "2016-02-22 12:32:01 UTC"
		{
			err := converter.AssignConverted(&value, date)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			err := converter.AssignConverted(&value, &date)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			err := converter.AssignConverted(&value, "1456144321")
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			unix := "1456144321.0"
			err := converter.AssignConverted(&value, unix)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			unix := 1456144321.0
			err := converter.AssignConverted(&value, unix)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}
		{
			unix := 1456144321.0
			err := converter.AssignConverted(&value, &unix)
			assert.Nil(t, err)
			assert.Equal(t, 1456144321, int(value.Unix()))
		}

		{
			date := "2016/02/22 12:32:01 UTC"
			err := converter.AssignConverted(&value, date)
			assert.NotNil(t, err, "invalid date format")
		}
	}

	{

		type A struct {
			Id   int
			Name string
			A    []string
		}

		aMap := map[string]interface{}{
			"Id":   1,
			"Name": "abc",
			"A":    []string{"a", "b"},
		}
		a := A{}
		err := converter.AssignConverted(&a, aMap)
		assert.Nil(t, err)
		assert.Equal(t, 1, a.Id)
		assert.Equal(t, "abc", a.Name)
		assert.Equal(t, 2, len(a.A))

	}

	{

		aMap := map[string]interface{}{
			"Id":   1,
			"Name": "abc",
			"A":    []string{"a", "b"},
		}
		{
			a := make(map[string]interface{})
			err := converter.AssignConverted(&a, aMap)
			assert.Nil(t, err)
			assert.Equal(t, 1, a["Id"])
			assert.Equal(t, "abc", a["Name"])
		}
		{
			a := make(map[string]interface{})
			err := converter.AssignConverted(&a, &aMap)
			assert.Nil(t, err)
			assert.Equal(t, 1, a["Id"])
			assert.Equal(t, "abc", a["Name"])
		}
	}

	{

		aSlice := []interface{}{1, 2, 3}
		target := make([]int, 0)
		err := converter.AssignConverted(&target, aSlice)
		assert.Nil(t, err)
		assert.Equal(t, 1, target[0])
		assert.Equal(t, 3, len(target))
	}
	{
		aSlice := []interface{}{1, 2, 3}
		target := make([]int, 0)
		err := converter.AssignConverted(&target, aSlice)
		assert.Nil(t, err)
		assert.Equal(t, 1, target[0])
		assert.Equal(t, 3, len(target))
	}

}

func TestAsString(t *testing.T) {
	assert.Equal(t, "abc", toolbox.AsString(([]byte)("abc")))
	assert.Equal(t, "123", toolbox.AsString("123"))
	var aInt uint = 1
	assert.Equal(t, "1", toolbox.AsString(aInt))
	type S struct {
		Id int
	}
	assert.Equal(t, "&{1}", toolbox.AsString(&S{1}))

}

func TestAsFloat(t *testing.T) {
	assert.Equal(t, 1.1, toolbox.AsFloat(1.1))
	assert.Equal(t, 0.0, toolbox.AsFloat("abc"))
}

func TestAsBoolean(t *testing.T) {
	assert.False(t, toolbox.AsBoolean(1.1))
	assert.True(t, toolbox.AsBoolean("true"))
}

func TestAsInt(t *testing.T) {
	assert.Equal(t, 1, toolbox.AsInt(1.1))
	assert.Equal(t, 0, toolbox.AsInt("avc"))
}

func TestDiscoverValueAndKind(t *testing.T) {
	{
		value, kind := toolbox.DiscoverValueAndKind("true")
		assert.Equal(t, true, value)
		assert.Equal(t, reflect.Bool, kind)
	}
	{
		value, kind := toolbox.DiscoverValueAndKind("abc")
		assert.Equal(t, "abc", value)
		assert.Equal(t, reflect.String, kind)
	}
	{
		value, kind := toolbox.DiscoverValueAndKind("3.4")
		assert.Equal(t, 3.4, value)
		assert.Equal(t, reflect.Float64, kind)
	}
	{
		value, kind := toolbox.DiscoverValueAndKind("3")
		assert.Equal(t, 3, value)
		assert.Equal(t, reflect.Int, kind)
	}
	{
		value, kind := toolbox.DiscoverValueAndKind("")
		assert.Nil(t, value)
		assert.Equal(t, reflect.Invalid, kind)

	}
}

func TestDiscoverCollectionValuesAndKind(t *testing.T) {
	{
		values, kind := toolbox.DiscoverCollectionValuesAndKind([]interface{}{
			1,
			2.3,
			"abc",
		})
		assert.Equal(t, reflect.String, kind)
		assert.Equal(t, "1", values[0])
		assert.Equal(t, "2.3", values[1])
		assert.Equal(t, "abc", values[2])

	}

	{
		values, kind := toolbox.DiscoverCollectionValuesAndKind([]interface{}{
			1,
			2.3,
		})
		assert.Equal(t, reflect.Float64, kind)
		assert.Equal(t, 1.0, values[0])
		assert.Equal(t, 2.3, values[1])

	}

	{
		values, kind := toolbox.DiscoverCollectionValuesAndKind([]interface{}{
			"true",
			false,
		})
		assert.Equal(t, reflect.Bool, kind)
		assert.Equal(t, true, values[0])
		assert.Equal(t, false, values[1])

	}

}

func TestDiscoverCollectionValueType(t *testing.T) {
	{
		var input = []string{"3.2", "1.2"}
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

func TestUnwrapValue(t *testing.T) {
	type S struct {
		F1 int
		F2 float64
		F3 string
		F4 uint
	}
	s := S{1, 1.1, "a", uint(1)}
	sValue := reflect.ValueOf(s)
	{
		fieldValue := sValue.FieldByName("F1")
		value := toolbox.UnwrapValue(&fieldValue)
		assert.Equal(t, 1, value)
	}
	{
		fieldValue := sValue.FieldByName("F2")
		value := toolbox.UnwrapValue(&fieldValue)
		assert.Equal(t, 1.1, value)
	}
	{
		fieldValue := sValue.FieldByName("F3")
		value := toolbox.UnwrapValue(&fieldValue)
		assert.Equal(t, "a", value)
	}

}
