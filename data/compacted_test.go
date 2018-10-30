package data

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
)

func TestNewCollection(t *testing.T) {
	collection := NewCompactedSlice(true, true)

	collection.Add(map[string]interface{}{
		"f1":  1,
		"f12": 1,
		"f15": 1,
		"f20": 1,
		"f11": nil,
		"f13": "",
	})

	collection.Add(map[string]interface{}{
		"f1":  1,
		"f32": 1,
		"f35": 1,
		"f30": 1,
		"f31": nil,
		"f33": "",
		"f11": 0,
		"f36": 0.0,
	})

	var actual = []map[string]interface{}{}
	collection.Range(func(data interface{}) (bool, error) {
		actual = append(actual, toolbox.AsMap(data))
		return true, nil
	})
	assert.Equal(t, 2, len(actual))
	assert.Equal(t, map[string]interface{}{
		"f1":  1,
		"f12": 1,
		"f15": 1,
		"f20": 1,
	}, actual[0])

	assert.Equal(t, map[string]interface{}{
		"f1":  1,
		"f32": 1,
		"f35": 1,
		"f30": 1,
	}, actual[1])
}

func Test_optimizedStorage(t *testing.T) {
	collection := NewCompactedSlice(true, true)
	var data = []interface{}{nil, nil, nil, "123", nil, nil, "abc", 12, nil, nil, nil, "a"}
	var compressed = []interface{}{nilGroup(3), "123", nilGroup(2), "abc", 12, nilGroup(3), "a"}
	var optimized = collection.compress(data)
	assert.EqualValues(t, compressed, optimized)
	collection.fields = make([]*field, 12)
	var uncompressed = make([]interface{}, len(collection.fields))
	collection.uncompress(compressed, uncompressed)
	assert.EqualValues(t, data, uncompressed)
}
