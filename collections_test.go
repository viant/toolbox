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
	"strings"
)

func TestIndexSlice(t *testing.T) {

	{
		type Foo struct {
			id   int
			name string
		}

		var fooCollection = []Foo{Foo{1, "A"}, Foo{2, "B"} }
		var indexedMap = make(map[int]Foo)
		toolbox.IndexSlice(fooCollection, indexedMap, func(foo Foo) int {
			return foo.id
		})
		assert.Equal(t, "A", indexedMap[1].name)
	}

	{
		aSlice := []string{"a", "c"}
		aMap :=make(map[string]int)
		index :=0
		toolbox.SliceToMap(aSlice, aMap, toolbox.CopyStringValueProvider, func(s string) int {
			index++
			return index
		})
		assert.Equal(t, 2, len(aMap))

	}
}

func TestProcessSlice(t *testing.T) {
	aSlice := []interface{}{
		"abc", "def", "cyz", "adc",
	}
	count := 0
	toolbox.ProcessSlice(aSlice, func(item interface{}) bool {
		count++
		return true
	})

	assert.Equal(t, 4, count)
}

func TestMakeMapFromSlice(t *testing.T) {
	type Foo struct {
		id   int
		name string
	}

	var fooCollection = []Foo{Foo{1, "A"}, Foo{2, "B"} }
	var testMap = make(map[int]string)
	toolbox.SliceToMap(fooCollection, testMap, func(foo Foo) int {
		return foo.id
	},
		func(foo Foo) string {
			return foo.name
		})

	assert.Equal(t, "A", testMap[1])
	assert.Equal(t, "B", testMap[2])

}

func TestProcess2DSliceInBatches(t *testing.T) {
	slice := [][]interface{}{
		[]interface{}{1, 2, 3},
		[]interface{}{4, 5, 7},
		[]interface{}{7, 8, 9},
		[]interface{}{10, 11, 12},
		[]interface{}{13, 14, 15},
		[]interface{}{16, 17, 18},
		[]interface{}{19, 20, 21},
	}

	actualItemCount := 0
	toolbox.Process2DSliceInBatches(slice, 2, func(item [][]interface{}) {
		actualItemCount = actualItemCount + len(item)
	})
	assert.Equal(t, 7, actualItemCount)
}

func TestAppendToSlice(t *testing.T) {
	source := []interface{}{
		"abc", "def", "cyz",
	}
	var target = make([]string, 0)
	toolbox.CopySliceElements(source, &target)
	assert.Equal(t, 3, len(target))
	for i := 0; i < len(source); i++ {
		assert.Equal(t, source[i], target[i])
	}
}

func TestFilterSliceElements(t *testing.T) {
	source := []interface{}{
		"abc", "def", "cyz", "adc",
	}
	var target = make([]string, 0)
	//filter all elements starting with a
	toolbox.FilterSliceElements(source, func(item string) bool {
		return strings.HasPrefix(item, "a")
	}, &target)
	assert.Equal(t, 2, len(target))
	assert.Equal(t, "abc", target[0])
	assert.Equal(t, "adc", target[1])

}

func TestHasSliceAnyElements(t *testing.T) {
	source := []interface{}{
		"abc", "def", "cyz", "adc",
	}
	assert.True(t, toolbox.HasSliceAnyElements(source, "cyz"))
	assert.False(t, toolbox.HasSliceAnyElements(source, "cyze"))
	assert.True(t, toolbox.HasSliceAnyElements(source, "cyze", "cyz"))
}

func TestMapKeysToSlice(t *testing.T) {
	m := map[string]int{
		"abc":1,
		"efg":2,
	}
	var keys = make([]string, 0)
	toolbox.MapKeysToSlice(m, &keys)
	assert.Equal(t, 2, len(keys))
}

func TestMapKeysToStringSlice(t *testing.T) {
	m := map[string]int{
		"abc":1,
		"efg":2,
	}
	slice := toolbox.MapKeysToStringSlice(m)
	assert.Equal(t, 2, len(slice))
}

func TestCopyMapEntries(t *testing.T) {
	type Foo struct{ id int; name string }
	source := map[interface{}]interface{}{
		1: Foo{1, "A"},
		2: Foo{2, "B"},
	}
	var target = make(map[int]Foo)

	toolbox.CopyMapEntries(source, target)
	assert.Equal(t, 2, len(target))
	assert.Equal(t, "B", target[2].name)
}

func TestIndexMultimap(t *testing.T) {
	type Product struct{ vendor, name string }
	products := []Product{
		Product{"Vendor1", "Product1"},
		Product{"Vendor2", "Product2"},
		Product{"Vendor1", "Product3"},
		Product{"Vendor1", "Product4"},
	}

	productsByVendor := make(map[string][]Product)
	toolbox.GroupSliceElements(products, productsByVendor, func(product Product) string {
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
		Product{"Vendor1", "Product1", 1},
		Product{"Vendor2", "Product2", 2},
		Product{"Vendor1", "Product3", 3},
		Product{"Vendor1", "Product4", 4},
	}

	productsByVendor := make(map[string][]int)
	toolbox.SliceToMultimap(products, productsByVendor, func(product Product) string {
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
		Product{"Vendor1", "Product1"},
		Product{"Vendor2", "Product2"},
		Product{"Vendor1", "Product3"},
		Product{"Vendor1", "Product4"},
	}
	var vendors=make([]string, 0)
	toolbox.TransformSlice(products, &vendors, func(product Product) string {
		return product.vendor
	})
	assert.Equal(t, 4, len(vendors))
	assert.Equal(t, "Vendor1", vendors[3])
}