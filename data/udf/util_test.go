package udf

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"testing"
)

func Test_Length(t *testing.T) {
	{
		value, err := Length(4.3, nil)
		assert.Nil(t, err)
		assert.Equal(t, 0, value)
	}
	{
		value, err := Length("abcd", nil)
		assert.Nil(t, err)
		assert.Equal(t, 4, value)
	}
	{
		value, err := Length(map[int]int{
			2: 3,
			1: 1,
			6: 3,
		}, nil)
		assert.Nil(t, err)
		assert.Equal(t, 3, value)
	}
	{
		value, err := Length([]int{1, 2, 3}, nil)
		assert.Nil(t, err)
		assert.Equal(t, 3, value)
	}
}

func Test_Keys(t *testing.T) {

	{
		var keys, err = Keys(map[string]interface{}{}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, keys)
	}
	{
		var keys, err = Keys("{\"abc\":1}", nil)
		assert.Nil(t, err)
		assert.EqualValues(t, []interface{}{"abc"}, keys)
	}
}

func Test_Values(t *testing.T) {

	{
		var keys, err = Values(map[string]interface{}{}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, keys)
	}
	{
		var keys, err = Values("{\"abc\":1}", nil)
		assert.Nil(t, err)
		assert.EqualValues(t, []interface{}{1}, keys)
	}
}

func Test_Join(t *testing.T) {
	{
		var joined, err = Join([]interface{}{
			[]interface{}{1, 2, 3},
			",",
		}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, joined)
		assert.EqualValues(t, "1,2,3", joined)
	}
}

func Test_Split(t *testing.T) {
	{
		var joined, err = Split([]interface{}{
			"abc , zc",
			",",
		}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, joined)
		assert.EqualValues(t, []string{"abc", "zc"}, joined)
	}
}

func Test_IndexOf(t *testing.T) {
	{
		index, _ := IndexOf([]interface{}{"this is test", "is"}, nil)
		assert.EqualValues(t, 2, index)
	}
	{
		index, _ := IndexOf([]interface{}{[]string{"this", "is", "test"}, "is"}, nil)
		assert.EqualValues(t, 1, index)
	}
}

func Test_Base64Decode(t *testing.T) {
	decoded, _ := Base64DecodeText("IkhlbGxvIFdvcmxkIg==", nil)
	assert.EqualValues(t, `"Hello World"`, decoded)

}

func TestTrimSpace(t *testing.T) {
	trimmed, _ := TrimSpace(" erer ", nil)
	assert.EqualValues(t, `erer`, trimmed)

}

func TestSum(t *testing.T) {
	{ //sum slice keys
		var aMap = data.NewMap()
		var collection = data.NewCollection()
		collection.Push(map[string]interface{}{
			"amount": 2,
		})
		collection.Push(map[string]interface{}{
			"amount": 12,
		})
		aMap.SetValue("node1.obj", collection)
		total, err := Sum("node1/obj/*/amount", aMap)
		assert.Nil(t, err)
		assert.Equal(t, 14, total)
	}
	{ //sum map keys
		var aMap = data.NewMap()
		aMap.SetValue("node1.obj.k1.amount", 1)
		aMap.SetValue("node1.obj.k2.amount", 2)
		aMap.SetValue("node1.obj.k3.amount", 3)
		total, err := Sum("node1/obj/*/amount", aMap)
		assert.Nil(t, err)
		assert.Equal(t, 6, total)
	}
}

func TestCount(t *testing.T) {
	{ //sum slice keys
		var aMap = data.NewMap()
		var collection = data.NewCollection()
		collection.Push(map[string]interface{}{
			"amount": 2,
		})
		collection.Push(map[string]interface{}{
			"amount": 12,
		})
		aMap.SetValue("node1.obj", collection)
		total, err := Count("node1/obj/*/amount", aMap)
		assert.Nil(t, err)
		assert.Equal(t, 2, total)
	}
	{ //sum map keys
		var aMap = data.NewMap()
		aMap.SetValue("node1.obj.k1.amount", 1)
		aMap.SetValue("node1.obj.k2.amount", 2)
		aMap.SetValue("node1.obj.k3.amount", 3)
		total, err := Count("node1/obj/*/amount", aMap)
		assert.Nil(t, err)
		assert.Equal(t, 3, total)
	}
}

func TestSelect(t *testing.T) {
	{ //sum slice keys
		var aMap = data.NewMap()
		var collection = data.NewCollection()
		collection.Push(map[string]interface{}{
			"amount": 2,
			"id":     2,
			"name":   "p1",
			"vendor": "v1",
		})
		collection.Push(map[string]interface{}{
			"amount": 12,
			"id":     3,
			"name":   "p2",
			"vendor": "v2",
		})
		aMap.SetValue("node1.obj", collection)

		records, err := Select([]interface{}{"node1/obj/*", "id", "name:product"}, aMap)
		assert.Nil(t, err)
		assert.Equal(t, []interface{}{
			map[string]interface{}{
				"id":      2,
				"product": "p1",
			},
			map[string]interface{}{
				"id":      3,
				"product": "p2",
			},
		}, records)

	}

}

func TestRand(t *testing.T) {
	{
		randValue, err := Rand(nil, nil)
		assert.Nil(t, err)
		floatValue, err := toolbox.ToFloat(randValue)
		assert.Nil(t, err)
		assert.True(t, toolbox.IsFloat(randValue) && floatValue >= 0.0 && floatValue < 1.0)
	}
	{
		randValue, err := Rand([]interface{}{2, 15}, nil)
		assert.Nil(t, err)
		intValue, err := toolbox.ToInt(randValue)
		assert.Nil(t, err)
		assert.True(t, toolbox.IsInt(randValue) && intValue >= 2 && intValue < 15)
	}
}

func TestConcat(t *testing.T) {
	{
		result, err := Concat([]interface{}{"a", "b", "c"}, nil)
		assert.Nil(t, err)
		assert.EqualValues(t, "abc", result)
	}

	{
		result, err := Concat([]interface{}{[]interface{}{"a", "b"}, "c"}, nil)
		assert.Nil(t, err)
		assert.EqualValues(t, []interface{}{"a", "b", "c"}, result)
	}

}

func TestMerge(t *testing.T) {
	{
		result, err := Merge([]interface{}{map[string]interface{}{
			"k1": 1,
		},
			map[string]interface{}{
				"k2": 2,
			},
		}, nil)
		assert.Nil(t, err)
		assert.EqualValues(t, map[string]interface{}{
			"k1": 1,
			"k2": 2,
		}, result)
	}
}
