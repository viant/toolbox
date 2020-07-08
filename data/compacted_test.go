package data

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
	"time"
)

func TestNewCollection(t *testing.T) {
	if ! canRun64BitArch() {
		t.Skip()
	}

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
	err := collection.Range(func(data interface{}) (bool, error) {
		actual = append(actual, toolbox.AsMap(data))
		return true, nil
	})
	assert.Nil(t, err)
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
	if ! canRun64BitArch() {
		t.Skip()
	}
	collection := NewCompactedSlice(true, true)
	var data = []interface{}{nil, nil, nil, "123", nil, nil, "abc", 12, nil, nil, nil, "a"}
	var compressed = []interface{}{nilGroup(3), "123", nilGroup(2), "abc", 12, nilGroup(3), "a"}
	var optimized = collection.compress(data)
	assert.EqualValues(t, compressed, optimized)
	collection.fields = make([]*Field, 12)
	var uncompressed = make([]interface{}, len(collection.fields))
	collection.uncompress(compressed, uncompressed)
	assert.EqualValues(t, data, uncompressed)
}

func TestCompactedSlice_SortedRange(t *testing.T) {
	if ! canRun64BitArch() {
		t.Skip()
	}
	var useCases = []struct {
		description string
		data        []map[string]interface{}
		expected    []interface{}
		indexBy     []string
		hasError    bool
	}{
		{
			description: "int sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   10,
					"name": "name 10",
				},
				{
					"id":   3,
					"name": "name 3",
				},
				{
					"id":   1,
					"name": "name 1",
				},
				{
					"id":   2,
					"name": "name 2",
				},
			},
			expected: []interface{}{
				1, 2, 3, 10,
			},
		},
		{
			description: "float sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   10.0,
					"name": "name 10",
				},
				{
					"id":   3.1,
					"name": "name 3",
				},
				{
					"id":   1.2,
					"name": "name 1",
				},
				{
					"id":   2.2,
					"name": "name 2",
				},
			},
			expected: []interface{}{
				1.2, 2.2, 3.1, 10.0,
			},
		},
		{
			description: "string sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   "010",
					"name": "name 10",
				},
				{
					"id":   "003",
					"name": "name 3",
				},
				{
					"id":   "001",
					"name": "name 1",
				},
				{
					"id":   "022",
					"name": "name 2",
				},
			},
			expected: []interface{}{
				"001", "003", "010", "022",
			},
		},
		{
			description: "combined index sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   1,
					"u":    1,
					"name": "name 10",
				},
				{
					"id":   3,
					"u":    2,
					"name": "name 3",
				},
				{
					"id":   2,
					"u":    2,
					"name": "name 1",
				},
				{
					"id":   4,
					"u":    6,
					"name": "name 2",
				},
			},
			expected: []interface{}{
				1, 2, 3, 4,
			},
		},
		{
			description: "missing Field",
			indexBy:     []string{"field1"},
			data: []map[string]interface{}{
				{
					"id":   1,
					"u":    1,
					"name": "name 10",
				},
			},
			hasError: true,
		},
		{
			description: "unsupported index type Field",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   time.Now(),
					"u":    1,
					"name": "name 10",
				},
			},
			hasError: true,
		},
	}

	for _, useCase := range useCases {
		collection := NewCompactedSlice(true, true)
		var actual = make([]interface{}, 0)
		for _, item := range useCase.data {
			collection.Add(item)
		}
		err := collection.SortedRange(useCase.indexBy, func(item interface{}) (b bool, e error) {
			record := toolbox.AsMap(item)
			actual = append(actual, record[useCase.indexBy[0]])
			return true, nil
		})
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, useCase.expected, actual, useCase.description)
	}

}

func TestCompactedSlice_SortedIterator(t *testing.T) {
	if ! canRun64BitArch() {
		t.Skip()
	}
	var useCases = []struct {
		description string
		data        []map[string]interface{}
		expected    []interface{}
		indexBy     []string
		hasError    bool
	}{
		{
			description: "int sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   10,
					"name": "name 10",
				},
				{
					"id":   3,
					"name": "name 3",
				},
				{
					"id":   1,
					"name": "name 1",
				},
				{
					"id":   2,
					"name": "name 2",
				},
			},
			expected: []interface{}{
				1, 2, 3, 10,
			},
		},
		{
			description: "float sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   10.0,
					"name": "name 10",
				},
				{
					"id":   3.1,
					"name": "name 3",
				},
				{
					"id":   1.2,
					"name": "name 1",
				},
				{
					"id":   2.2,
					"name": "name 2",
				},
			},
			expected: []interface{}{
				1.2, 2.2, 3.1, 10.0,
			},
		},
		{
			description: "string sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   "010",
					"name": "name 10",
				},
				{
					"id":   "003",
					"name": "name 3",
				},
				{
					"id":   "001",
					"name": "name 1",
				},
				{
					"id":   "022",
					"name": "name 2",
				},
			},
			expected: []interface{}{
				"001", "003", "010", "022",
			},
		},
		{
			description: "combined index sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   1,
					"u":    1,
					"name": "name 10",
				},
				{
					"id":   3,
					"u":    2,
					"name": "name 3",
				},
				{
					"id":   2,
					"u":    2,
					"name": "name 1",
				},
				{
					"id":   4,
					"u":    6,
					"name": "name 2",
				},
			},
			expected: []interface{}{
				1, 2, 3, 4,
			},
		},
		{
			description: "missing Field",
			indexBy:     []string{"field1"},
			data: []map[string]interface{}{
				{
					"id":   1,
					"u":    1,
					"name": "name 10",
				},
			},
			hasError: true,
		},
		{
			description: "unsupported index type Field",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   time.Now(),
					"u":    1,
					"name": "name 10",
				},
			},
			hasError: true,
		},
	}

	for _, useCase := range useCases {
		collection := NewCompactedSlice(true, true)
		var actual = make([]interface{}, 0)
		for _, item := range useCase.data {
			collection.Add(item)
		}
		iterator, err := collection.SortedIterator(useCase.indexBy)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		var record map[string]interface{}
		for iterator.HasNext() {
			err = iterator.Next(&record)
			assert.Nil(t, err)
			actual = append(actual, record[useCase.indexBy[0]])
		}
		assert.EqualValues(t, useCase.expected, actual, useCase.description)
	}

}

func TestCompactedSlice_Iterator(t *testing.T) {
	if ! canRun64BitArch() {
		t.Skip()
	}
	var useCases = []struct {
		description string
		data        []map[string]interface{}
		expected    []interface{}
		indexBy     []string
		hasError    bool
	}{
		{
			description: "int sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   10,
					"name": "name 10",
				},
				{
					"id":   3,
					"name": "name 3",
				},
				{
					"id":   1,
					"name": "name 1",
				},
			},
			expected: []interface{}{
				10, 3, 1,
			},
		},
		{
			description: "float sorting",
			indexBy:     []string{"id"},
			data: []map[string]interface{}{
				{
					"id":   10.0,
					"name": "name 10",
				},
				{
					"id":   3.1,
					"name": "name 3",
				},
				{
					"id":   2.2,
					"name": "name 2",
				},
			},
			expected: []interface{}{
				10.0, 3.1, 2.2,
			},
		},
	}

	for _, useCase := range useCases {
		collection := NewCompactedSlice(true, true)
		var actual = make([]interface{}, 0)
		for _, item := range useCase.data {
			collection.Add(item)
		}
		iterator := collection.Iterator()

		var record map[string]interface{}
		for iterator.HasNext() {
			err := iterator.Next(&record)
			assert.Nil(t, err)
			actual = append(actual, record[useCase.indexBy[0]])
		}
		assert.EqualValues(t, useCase.expected, actual, useCase.description)
	}

}


func TestCompactedSlice_MarshalJSON(t *testing.T) {
	if ! canRun64BitArch() {
		t.Skip()
	}
	var useCases = []struct {
		description string
		data        []map[string]interface{}
		hasError    bool
	}{
		{
			description: "array marshaling",
			data: []map[string]interface{}{
				{
					"id":   float64(10),
					"name": "name 10",
				},
				{
					"id":   float64(3),
					"name": "name 3",
				},
				{
					"id":   float64(1),
					"name": "name 1",
				},
			},

		},

	}

	for _, useCase := range useCases {
		collection := NewCompactedSlice(true, true)

		for _, item := range useCase.data {
			collection.Add(item)
		}
		rawJSON, err := json.Marshal(collection)
		if ! assert.Nil(t, err, useCase.description) {
			continue
		}
		actual := []map[string]interface{}{}
		json.Unmarshal(rawJSON, &actual)
		assert.EqualValues(t, useCase.data, actual)




	}

}


func canRun64BitArch() bool {
	isNot64BitArch := 32 << uintptr(^uintptr(0)>>63) < 64
	return ! isNot64BitArch
}