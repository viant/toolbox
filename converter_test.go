package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"reflect"
	"testing"
	"time"
)

func TestConverter(t *testing.T) {
	converter := toolbox.NewColumnConverter(toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z"))

	{

		type A1 struct {
			K1 int
		}
		type A2 struct {
			K2 int
		}

		type C struct {
			*A1
			*A2
			K3 int
		}

		aMap := map[string]interface{}{
			"K1": 1,
			"K2": 20,
			"K3": 30,
		}
		c := C{}
		err := converter.AssignConverted(&c, aMap)
		assert.Nil(t, err)
		assert.Equal(t, 1, c.K1)
		assert.Equal(t, 20, c.K2)
		assert.Equal(t, 30, c.K3)

	}

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
		var testData = []interface{}{int(102), int64(102), float64(102), float32(102), "102"}
		for _, item := range testData {
			err := converter.AssignConverted(&value, item)
			if assert.Nil(t, err) {
				if assert.NotNil(t, value) {
					assert.Equal(t, float64(102), *value)
				}
			}
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
			if assert.Nil(t, err) {
				assert.Equal(t, 1456144321, int(value.Unix()))
			}
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
			var unixNano = "1513271472347277824"
			err := converter.AssignConverted(&value, unixNano)
			assert.Nil(t, err)
			assert.Equal(t, 1513271472, int(value.Unix()))
		}

		{
			var unixNano = "1456144321001"
			err := converter.AssignConverted(&value, unixNano)
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



		//{
		//	unix := 1668069210749
		//	err := converter.AssignConverted(&value, &unix)
		//	assert.Nil(t, err)
		//	assert.Equal(t, 1668069210, int(value.Unix()))
		//}
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

func Test_Converter_SliceToMap(t *testing.T) {

	//KeyValue represents sorted map entry
	type KeyValue struct {
		Key, Value interface{}
	}

	{
		converter := toolbox.NewColumnConverter(toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z"))
		var slice = []*KeyValue{
			{Key: "k1", Value: 1},
			{Key: "k2", Value: 2},
		}

		var aMap = make(map[string]interface{})
		err := converter.AssignConverted(&aMap, slice)
		assert.Nil(t, err)
		assert.EqualValues(t, map[string]interface{}{
			"k1": 1,
			"k2": 2,
		}, aMap)
	}
	{
		converter := toolbox.NewColumnConverter(toolbox.DateFormatToLayout("yyyy-MM-dd hh:mm:ss z"))
		var slice = []map[string]interface{}{
			{"Key": "k1", "Value": 1},
			{"Key": "k2", "Value": 2},
		}
		var aMap = make(map[string]interface{})
		err := converter.AssignConverted(&aMap, slice)
		assert.Nil(t, err)
		assert.EqualValues(t, map[string]interface{}{
			"k1": 1,
			"k2": 2,
		}, aMap)
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

	{
		var bytes = []uint8{
			34,
			72,
			101,
			108,
			108,
			111,
			32,
			87,
			111,
			114,
			108,
			100,
			34,
		}
		assert.EqualValues(t, `"Hello World"`, toolbox.AsString(bytes))
	}

	{
		var bytes = []interface{}{
			uint8(34),
			uint8(72),
			uint8(101),
			uint8(108),
			uint8(108),
			uint8(111),
			uint8(32),
			uint8(87),
			uint8(111),
			uint8(114),
			uint8(108),
			uint8(100),
			uint8(34),
		}
		assert.EqualValues(t, `"Hello World"`, toolbox.AsString(bytes))
	}

}

func TestAsFloat(t *testing.T) {
	assert.Equal(t, 1.1, toolbox.AsFloat(1.1))
	assert.Equal(t, 0.0, toolbox.AsFloat("abc"))
}

func TestAsBoolean(t *testing.T) {
	assert.False(t, toolbox.AsBoolean(1.1))
	assert.True(t, toolbox.AsBoolean("true"))
	assert.True(t, toolbox.AsBoolean(0x1))
	assert.False(t, toolbox.AsBoolean(0x0))

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

func TestConverter_AsInt(t *testing.T) {
	intValue := toolbox.AsInt("5.638679022673832e+18")
	assert.True(t, intValue > 0)

}

func TestConvertedMapFromStruct(t *testing.T) {
	var aStruct = struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string
	}{1, "test", "desc"}
	converter := toolbox.NewConverter("", "json")
	var target = make(map[string]interface{})
	err := converter.AssignConverted(&target, aStruct)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]interface{}{
		"id":          1,
		"name":        "test",
		"Description": "desc",
	}, target)
}

func TestConvertedSliceToMapError(t *testing.T) {
	aSlice := []map[string]interface{}{
		{
			"id":   1,
			"name": 111,
		},
		{
			"id":   2,
			"name": 222,
		},
	}

	var aMap = make(map[string]interface{})

	converter := toolbox.NewConverter("", "json")
	err := converter.AssignConverted(&aMap, aSlice)
	assert.NotNil(t, err)
}
