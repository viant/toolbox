package toolbox

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestIndexSliceAsync(t *testing.T) {

	{
		type Foo struct {
			id   int
			name string
		}

		var fooCollection = []Foo{{1, "A"}, {2, "B"}}
		var indexedMap = make(map[int]Foo)
		IndexSliceAsync(fooCollection, indexedMap, func(foo Foo) int {
			return foo.id
		})
		assert.Equal(t, "A", indexedMap[1].name)
	}

	{
		aSlice := []string{"a", "c"}
		aMap := make(map[string]int)
		index := 0
		SliceToMapAsync(aSlice, aMap, CopyStringValueProvider, func(s string) int {
			index++
			return index
		})
		assert.Equal(t, 2, len(aMap))

	}
}

type sliceItem struct {
	Id int
}

func TestReverseSlice(t *testing.T) {

	{
		aSlice := []interface{}{
			"abc", "def", "cyz", "adc", "z",
		}

		ReverseSlice(aSlice)
		assert.Equal(t, []interface{}{"z", "adc", "cyz", "def", "abc"}, aSlice)
	}

	ReverseSlice(nil)
	{
		aSlice := []*sliceItem{
			{1}, {10},
		}

		ReverseSlice(aSlice)
		assert.Equal(t, []*sliceItem{
			{10}, {1},
		}, aSlice)
	}

}

func TestProcessSliceAsync(t *testing.T) {
	{
		aSlice := []interface{}{
			"abc", "def", "cyz", "adc",
		}
		count := 0
		ProcessSliceAsync(aSlice, func(item interface{}) bool {
			count++
			return true
		})

		assert.Equal(t, 4, count)
	}

	{
		aSlice := []string{
			"abc", "def", "cyz", "adc",
		}
		count := 0
		ProcessSliceAsync(aSlice, func(item interface{}) bool {
			count++
			return true
		})

		assert.Equal(t, 4, count)
	}
}

func TestProcessSliceWithIndexAsync(t *testing.T) {
	{
		aSlice := []interface{}{
			"abc", "def", "cyz", "adc",
		}
		count := 0
		ProcessSliceWithIndexAsync(aSlice, func(index int, item interface{}) bool {
			count += 1 //Test case changed due to index being async
			return true
		})

		assert.Equal(t, 4, count)
	}
	{
		aSlice := []string{
			"abc", "def", "cyz", "adc",
		}
		count := 0
		ProcessSliceWithIndexAsync(aSlice, func(index int, item interface{}) bool {
			count += 1
			return true
		})

		assert.Equal(t, 4, count)
	}
}

func TestMakeMapFromSliceAsync(t *testing.T) {
	type Foo struct {
		id   int
		name string
	}

	var fooCollection = []Foo{{1, "A"}, {2, "B"}}
	var testMap = make(map[int]string)
	SliceToMapAsync(fooCollection, testMap, func(foo Foo) int {
		return foo.id
	},
		func(foo Foo) string {
			return foo.name
		})

	assert.Equal(t, "A", testMap[1])
	assert.Equal(t, "B", testMap[2])

}

func TestSliceToMapAsync(t *testing.T) {
	aSlice := []string{"a", "c"}
	aMap := make(map[string]bool)

	SliceToMapAsync(aSlice, aMap, func(s string) string {
		return s
	}, func(s string) bool {
		return true
	})
	assert.Equal(t, 2, len(aMap))

}

func TestProcess2DSliceInBatches(t *testing.T) {
	slice := [][]interface{}{
		{1, 2, 3},
		{4, 5, 7},
		{7, 8, 9},
		{10, 11, 12},
		{13, 14, 15},
		{16, 17, 18},
		{19, 20, 21},
	}

	actualItemCount := 0
	Process2DSliceInBatches(slice, 2, func(item [][]interface{}) {
		actualItemCount = actualItemCount + len(item)
	})
	assert.Equal(t, 7, actualItemCount)
}

func TestCopySliceElements(t *testing.T) {
	{
		source := []interface{}{
			"abc", "def", "cyz",
		}
		var target = make([]string, 0)
		CopySliceElements(source, &target)
		assert.Equal(t, 3, len(target))
		for i := 0; i < len(source); i++ {
			assert.Equal(t, source[i], target[i])
		}
	}
	{
		source := []interface{}{
			1, 2, 3,
		}
		var target = make([]int, 0)
		CopySliceElements(source, &target)
		assert.Equal(t, 3, len(target))
		for i := 0; i < len(source); i++ {
			assert.Equal(t, source[i], target[i])
		}
	}
	{
		source := []interface{}{
			1, 2, 3,
		}
		var target = make([]interface{}, 0)
		CopySliceElements(source, &target)
		assert.Equal(t, 3, len(target))
		for i := 0; i < len(source); i++ {
			assert.Equal(t, source[i], target[i])
		}
	}

}

func TestFilterSliceElements(t *testing.T) {
	{
		source := []interface{}{
			1, 2, 3,
		}
		var target = make([]int, 0)
		//filter all elements starting with a
		FilterSliceElements(source, func(item int) bool {
			return item > 1
		}, &target)
		assert.Equal(t, 2, len(target))
		assert.Equal(t, 2, target[0])
		assert.Equal(t, 3, target[1])
	}

	{
		source := []interface{}{
			"abc", "def", "cyz", "adc",
		}
		var target = make([]string, 0)
		//filter all elements starting with a
		FilterSliceElements(source, func(item string) bool {
			return strings.HasPrefix(item, "a")
		}, &target)
		assert.Equal(t, 2, len(target))
		assert.Equal(t, "abc", target[0])
		assert.Equal(t, "adc", target[1])
	}

}

func TestHasSliceAnyElements(t *testing.T) {
	source := []interface{}{
		"abc", "def", "cyz", "adc",
	}
	assert.True(t, HasSliceAnyElements(source, "cyz"))
	assert.False(t, HasSliceAnyElements(source, "cyze"))
	assert.True(t, HasSliceAnyElements(source, "cyze", "cyz"))
}

func TestMapKeysToSlice(t *testing.T) {
	m := map[string]int{
		"abc": 1,
		"efg": 2,
	}
	var keys = make([]string, 0)
	MapKeysToSlice(m, &keys)
	assert.Equal(t, 2, len(keys))
}

func TestMapKeysToStringSlice(t *testing.T) {
	m := map[string]int{
		"abc": 1,
		"efg": 2,
	}
	slice := MapKeysToStringSlice(m)
	assert.Equal(t, 2, len(slice))
}

func TestCopyMapEntries(t *testing.T) {
	type Foo struct {
		id   int
		name string
	}
	source := map[interface{}]interface{}{
		1: Foo{1, "A"},
		2: Foo{2, "B"},
	}
	var target = make(map[int]Foo)

	CopyMapEntries(source, target)
	assert.Equal(t, 2, len(target))
	assert.Equal(t, "B", target[2].name)
}

func TestIndexMultimap(t *testing.T) {
	type Product struct{ vendor, name string }
	products := []Product{
		{"Vendor1", "Product1"},
		{"Vendor2", "Product2"},
		{"Vendor1", "Product3"},
		{"Vendor1", "Product4"},
	}

	productsByVendor := make(map[string][]Product)
	GroupSliceElements(products, productsByVendor, func(product Product) string {
		return product.vendor
	})
	assert.Equal(t, 2, len(productsByVendor))
	assert.Equal(t, 3, len(productsByVendor["Vendor1"]))
	assert.Equal(t, "Product4", productsByVendor["Vendor1"][2].name)

}

func TestSliceToMultiMap(t *testing.T) {
	type Product struct {
		vendor, name string
		productId    int
	}

	products := []Product{
		{"Vendor1", "Product1", 1},
		{"Vendor2", "Product2", 2},
		{"Vendor1", "Product3", 3},
		{"Vendor1", "Product4", 4},
	}

	productsByVendor := make(map[string][]int)
	SliceToMultimap(products, productsByVendor, func(product Product) string {
		return product.vendor
	},
		func(product Product) int {
			return product.productId
		})

	assert.Equal(t, 2, len(productsByVendor))
	assert.Equal(t, 3, len(productsByVendor["Vendor1"]))
	assert.Equal(t, 4, productsByVendor["Vendor1"][2])

}

func TestTransformSlice(t *testing.T) {
	type Product struct{ vendor, name string }
	products := []Product{
		{"Vendor1", "Product1"},
		{"Vendor2", "Product2"},
		{"Vendor1", "Product3"},
		{"Vendor1", "Product4"},
	}
	var vendors = make([]string, 0)
	TransformSlice(products, &vendors, func(product Product) string {
		return product.vendor
	})
	assert.Equal(t, 4, len(vendors))
	assert.Equal(t, "Vendor1", vendors[3])
}

func TestMakeStringMap(t *testing.T) {
	aMap := MakeStringMap("a:1, b:2", ":", ",")
	assert.Equal(t, 2, len(aMap))
	assert.Equal(t, "1", aMap["a"])
	assert.Equal(t, "2", aMap["b"])
}

func TestMakeReverseStringMap(t *testing.T) {
	aMap := MakeReverseStringMap("a:1, b:2", ":", ",")
	assert.Equal(t, 2, len(aMap))
	assert.Equal(t, "a", aMap["1"])
	assert.Equal(t, "b", aMap["2"])
}

func TestSortStrings(t *testing.T) {
	sorted := SortStrings([]string{"z", "b", "c", "a"})
	assert.Equal(t, "a", sorted[0])
	assert.Equal(t, "z", sorted[3])

}

func TestJoinAsString(t *testing.T) {
	assert.Equal(t, "a,b", JoinAsString([]string{"a", "b"}, ","))
}

func TestSetSliceValue(t *testing.T) {

	{
		var aSlice = make([]string, 2)
		SetSliceValue(aSlice, 0, "abc")
		assert.Equal(t, "abc", aSlice[0])
		assert.Equal(t, "abc", GetSliceValue(aSlice, 0))
	}

	{
		var aSlice = make([]int, 2)
		SetSliceValue(aSlice, 0, 100)
		assert.Equal(t, 100, aSlice[0])
		assert.Equal(t, 100, GetSliceValue(aSlice, 0))
	}
	{
		var aSlice = make([]interface{}, 2)
		SetSliceValue(aSlice, 0, "a")
		assert.Equal(t, "a", aSlice[0])
		assert.Equal(t, "a", GetSliceValue(aSlice, 0))
	}
}

func TestTrueValueProvider(t *testing.T) {
	assert.True(t, TrueValueProvider(1))
}

func Test_DeleteEmptyKeys(t *testing.T) {
	aMap := map[string]interface{}{
		"k1": []int{},
		"k2": []int{1},
		"k3": "",
		"k40": map[interface{}]interface{}{
			"k1":  nil,
			1:     2,
			"k31": []map[string]interface{}{},
			"k41": []map[string]interface{}{
				{
					"z": 1,
				},
			},
		},
		"k5": map[string]interface{}{
			"k1": "",
			"10": 20,
		},
	}
	cloned := DeleteEmptyKeys(aMap)
	assert.Equal(t, map[string]interface{}{
		"k2": []interface{}{1},
		"k40": map[interface{}]interface{}{
			1: 2,
			"k41": []interface{}{
				map[string]interface{}{
					"z": 1,
				},
			},
		},
		"k5": map[string]interface{}{
			"10": 20,
		},
	}, cloned)
}

func TestIntersection(t *testing.T) {

	useCase1Actual := []string{}
	useCase2Actual := []int{}
	useCase3Actual := []float32{}

	var useCases = []struct {
		description string
		sliceA      interface{}
		sliceB      interface{}
		actual      interface{}
		expect      interface{}
		hasError    bool
	}{
		{
			description: "string slice intersection",
			sliceA:      []string{"a", "bc", "z", "eee"},
			sliceB:      []string{"a2", "bc", "5z", "eee"},
			actual:      &useCase1Actual,
			expect:      []string{"bc", "eee"},
		},
		{
			description: "int slice intersection",
			sliceA:      []int{1, 2, 3, 4},
			sliceB:      []int{3, 4, 5, 6},
			actual:      &useCase2Actual,
			expect:      []int{3, 4},
		},
		{
			description: "float slice intersection",
			sliceA:      []float32{1.1, 2.1, 3.1, 4.1},
			sliceB:      []float32{3.1, 4.1, 5.1, 6.1},
			actual:      &useCase3Actual,
			expect:      []float32{3.1, 4.1},
		},
	}

	for _, useCase := range useCases {
		err := Intersect(useCase.sliceA, useCase.sliceB, useCase.actual)
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actual := reflect.ValueOf(useCase.actual).Elem().Interface()
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}
