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

// Package toolbox - collection utilities
package toolbox

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

//TrueValueProvider is a function that returns true, it takes one parameters which ignores,
//This provider can be used to make map from slice like map[some type]bool
var TrueValueProvider = func(ignore interface{}) bool {
	return true
}

//CopyStringValueProvider is a function that returns passed in string
//This provider can be used to make map from slice like map[string]some type
var CopyStringValueProvider = func(source string) string {
	return source
}

//ProcessSlice iterates over any slice, it calls handler with each element unless handler returns false,
func ProcessSlice(slice interface{}, handler func(item interface{}) bool) {
	//The common cases with reflection for speed
	if aSlice, ok := slice.([]interface{}); ok {
		for _, item := range aSlice {
			handler(item)
		}
		return
	}
	//The common cases with reflection for speed
	if aSlice, ok := slice.([]string); ok {
		for _, item := range aSlice {
			handler(item)
		}
		return
	}
	sliceValue := DiscoverValueByKind(reflect.ValueOf(slice), reflect.Slice)
	for i := 0; i < sliceValue.Len(); i++ {
		if !handler(sliceValue.Index(i).Interface()) {
			break
		}
	}
}

//ProcessSliceWithIndex iterates over any slice, it calls handler with every index and item unless handler returns false
func ProcessSliceWithIndex(slice interface{}, handler func(index int, item interface{}) bool) {
	if aSlice, ok := slice.([]interface{}); ok {
		for i, item := range aSlice {
			handler(i, item)
		}
		return
	}
	if aSlice, ok := slice.([]string); ok {
		for i, item := range aSlice {
			handler(i, item)
		}
		return
	}
	sliceValue := DiscoverValueByKind(reflect.ValueOf(slice), reflect.Slice)
	for i := 0; i < sliceValue.Len(); i++ {
		if !handler(i, sliceValue.Index(i).Interface()) {
			break
		}
	}
}

//IndexSlice reads passed in slice and applies function that takes a slice item as argument to return a key value.
//passed in resulting map needs to match key type return by a key function, and accept slice item type as argument.
func IndexSlice(slice, resultingMap, keyFunction interface{}) {
	mapValue := DiscoverValueByKind(resultingMap, reflect.Map)
	ProcessSlice(slice, func(item interface{}) bool {
		result := CallFunction(keyFunction, item)
		mapValue.SetMapIndex(reflect.ValueOf(result[0]), reflect.ValueOf(item))
		return true
	})
}

//CopySliceElements appends elements from source slice into target
//This function comes handy if you want to copy from generic []interface{} slice to more specific slice like []string, if source slice element are of the same time
func CopySliceElements(sourceSlice, targetSlicePointer interface{}) {

	if aTargetSlicePointer, ok := targetSlicePointer.(*[]interface{}); ok {
		ProcessSlice(sourceSlice, func(item interface{}) bool {
			*(aTargetSlicePointer) = append(*aTargetSlicePointer, item)
			return true
		})
		return
	}

	if aTargetSlicePointer, ok := targetSlicePointer.(*[]string); ok {
		ProcessSlice(sourceSlice, func(item interface{}) bool {
			*(aTargetSlicePointer) = append(*aTargetSlicePointer, AsString(item))
			return true
		})
		return
	}
	AssertPointerKind(targetSlicePointer, reflect.Slice, "targetSlicePointer")
	sliceValue := reflect.ValueOf(targetSlicePointer).Elem()
	ProcessSlice(sourceSlice, func(item interface{}) bool {
		sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(item)))
		return true
	})
}

//TransformSlice appends transformed elements from source slice into target, transformer take as argument item of source slice and return value of target slice.
func TransformSlice(sourceSlice, targetSlicePointer, transformer interface{}) {
	AssertPointerKind(targetSlicePointer, reflect.Slice, "targetSlicePointer")
	sliceValue := reflect.ValueOf(targetSlicePointer).Elem()
	ProcessSlice(sourceSlice, func(item interface{}) bool {
		result := CallFunction(transformer, item)
		sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(result[0])))
		return true
	})
}

//FilterSliceElements copies elements from sourceSlice to targetSlice if predicate function returns true. Predicate function needs to accept source slice element type and return true.
func FilterSliceElements(sourceSlice interface{}, predicate interface{}, targetSlicePointer interface{}) {
	//The most common case witout reflection
	if aTargetSlicePointer, ok := targetSlicePointer.(*[]string); ok {
		aPredicate, ok := predicate.(func(item string) bool)
		if !ok {
			panic("Invalid predicate")
		}
		ProcessSlice(sourceSlice, func(item interface{}) bool {
			if aPredicate(AsString(item)) {
				*(aTargetSlicePointer) = append(*aTargetSlicePointer, AsString(item))
			}
			return true
		})
		return
	}
	AssertPointerKind(targetSlicePointer, reflect.Slice, "targetSlicePointer")
	slicePointerValue := reflect.ValueOf(targetSlicePointer).Elem()
	ProcessSlice(sourceSlice, func(item interface{}) bool {
		result := CallFunction(predicate, item)
		if AsBoolean(result[0]) {
			slicePointerValue.Set(reflect.Append(slicePointerValue, reflect.ValueOf(item)))
		}
		return true
	})
}

//HasSliceAnyElements checks if sourceSlice has any of passed in elements. This method iterates through elements till if finds the first match.
func HasSliceAnyElements(sourceSlice interface{}, elements ...interface{}) (result bool) {
	ProcessSlice(sourceSlice, func(item interface{}) bool {
		for _, element := range elements {
			if item == element {
				result = true
				return false
			}
		}
		return true
	})
	return result
}

//SliceToMap reads passed in slice to to apply the key and value function for each item. Result of these calls is placed in the resulting map.
func SliceToMap(sourceSlice, targetMap, keyFunction, valueFunction interface{}) {
	mapValue := DiscoverValueByKind(targetMap, reflect.Map)
	ProcessSlice(sourceSlice, func(item interface{}) bool {
		key := CallFunction(keyFunction, item)
		value := CallFunction(valueFunction, item)
		mapValue.SetMapIndex(reflect.ValueOf(key[0]), reflect.ValueOf(value[0]))
		return true
	})
}

//GroupSliceElements reads source slice and transfer all values returned by keyFunction to a slice in target map.
func GroupSliceElements(sourceSlice, targetMap, keyFunction interface{}) {
	mapValue := DiscoverValueByKind(targetMap, reflect.Map)
	mapValueType := mapValue.Type().Elem()
	ProcessSlice(sourceSlice, func(item interface{}) bool {
		result := CallFunction(keyFunction, item)
		keyValue := reflect.ValueOf(result[0])
		sliceForThisKey := mapValue.MapIndex(keyValue)
		if !sliceForThisKey.IsValid() {
			sliceForThisKeyPoiner := reflect.New(mapValueType)
			sliceForThisKey = sliceForThisKeyPoiner.Elem()
			mapValue.SetMapIndex(keyValue, sliceForThisKey)
		}
		mapValue.SetMapIndex(keyValue, reflect.Append(sliceForThisKey, reflect.ValueOf(item)))
		return true
	})
}

//SliceToMultimap reads source slice and transfer all values by valueFunction and returned by keyFunction to a slice in target map.
//Key and value function result type need to agree with target map type.
func SliceToMultimap(sourceSlice, targetMap, keyFunction, valueFunction interface{}) {
	mapValue := DiscoverValueByKind(targetMap, reflect.Map)
	mapValueType := mapValue.Type().Elem()
	ProcessSlice(sourceSlice, func(item interface{}) bool {
		keyResult := CallFunction(keyFunction, item)
		keyValue := reflect.ValueOf(keyResult[0])

		valueResult := CallFunction(valueFunction, item)
		value := reflect.ValueOf(valueResult[0])
		sliceForThisKey := mapValue.MapIndex(keyValue)
		if !sliceForThisKey.IsValid() {
			sliceForThisKeyPoiner := reflect.New(mapValueType)
			sliceForThisKey = sliceForThisKeyPoiner.Elem()
			mapValue.SetMapIndex(keyValue, sliceForThisKey)
		}
		mapValue.SetMapIndex(keyValue, reflect.Append(sliceForThisKey, value))
		return true
	})
}

//SetSliceValue sets value at slice index
func SetSliceValue(slice interface{}, index int, value interface{}) {
	if aSlice, ok := slice.([]string); ok {
		aSlice[index] = AsString(value)
		return
	}
	if aSlice, ok := slice.([]interface{}); ok {
		aSlice[index] = value
		return
	}
	sliceValue := DiscoverValueByKind(reflect.ValueOf(slice), reflect.Slice)
	sliceValue.Index(index).Set(reflect.ValueOf(value))
}

//GetSliceValue gets value from passed in index
func GetSliceValue(slice interface{}, index int) interface{} {
	if aSlice, ok := slice.([]string); ok {
		return aSlice[index]
	}
	if aSlice, ok := slice.([]interface{}); ok {
		return aSlice[index]
	}
	sliceValue := DiscoverValueByKind(reflect.ValueOf(slice), reflect.Slice)
	return sliceValue.Index(index).Interface()
}

//ProcessMap iterates over any map, it calls handler with every key, value pair unless handler returns false.
func ProcessMap(sourceMap interface{}, handler func(key, value interface{}) bool) {
	mapValue := DiscoverValueByKind(reflect.ValueOf(sourceMap), reflect.Map)
	for _, key := range mapValue.MapKeys() {
		value := mapValue.MapIndex(key)
		if !handler(key.Interface(), value.Interface()) {
			break
		}
	}
}

//CopyMapEntries appends map entry from source map to target map
func CopyMapEntries(sourceMap, targetMap interface{}) {
	targetMapValue := reflect.ValueOf(targetMap)
	if targetMapValue.Kind() == reflect.Ptr {
		targetMapValue = targetMapValue.Elem()
	}
	ProcessMap(sourceMap, func(key, value interface{}) bool {
		targetMapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		return true
	})
}

//MapKeysToSlice appends all map keys to targetSlice
func MapKeysToSlice(sourceMap interface{}, targetSlicePointer interface{}) {
	AssertPointerKind(targetSlicePointer, reflect.Slice, "targetSlicePointer")
	slicePointerValue := reflect.ValueOf(targetSlicePointer).Elem()
	ProcessMap(sourceMap, func(key, value interface{}) bool {
		slicePointerValue.Set(reflect.Append(slicePointerValue, reflect.ValueOf(key)))
		return true
	})
}

//MapKeysToStringSlice creates a string slice from sourceMap keys, keys do not need to be of a string type.
func MapKeysToStringSlice(sourceMap interface{}) []string {
	var keys = make([]string, 0)
	ProcessMap(&sourceMap, func(key interface{}, value interface{}) bool {
		keys = append(keys, AsString(key))
		return true
	})
	return keys
}

//Process2DSliceInBatches iterates over any 2 dimensional slice, it calls handler with batch.
func Process2DSliceInBatches(slice [][]interface{}, size int, handler func(batchedSlice [][]interface{})) {
	batchCount := (len(slice) / size) + 1
	fromIndex, toIndex := 0, 0

	for i := 0; i < batchCount; i++ {
		toIndex = size * (i + 1)
		isLastBatch := toIndex >= len(slice)
		if isLastBatch {
			toIndex = len(slice)
		}
		handler(slice[fromIndex:toIndex])
		fromIndex = toIndex

	}
}

//SortStrings creates a new copy of passed in slice and sorts it.
func SortStrings(source []string) []string {
	var result = make([]string, 0)
	result = append(result, source...)
	sort.Strings(result)
	return result
}

//JoinAsString joins all items of a slice, with separator, it takes any slice as argument,
func JoinAsString(slice interface{}, separator string) string {
	result := ""
	ProcessSlice(slice, func(item interface{}) bool {
		if len(result) > 0 {
			result = result + separator
		}
		result = fmt.Sprintf("%v%v", result, item)
		return true
	})
	return result
}

//MakeStringMap creates a mapstring]string from string,
func MakeStringMap(text string, valueSepartor string, itemSeparator string) map[string]string {
	var result = make(map[string]string)
	for _, item := range strings.Split(text, itemSeparator) {

		if len(item) == 0 {
			continue
		}
		keyValue := strings.SplitN(item, valueSepartor, 2)
		if len(keyValue) == 2 {
			result[strings.Trim(keyValue[0], " \t")] = strings.Trim(keyValue[1], " \n\t")
		}
	}
	return result
}

//MakeReverseStringMap creates a mapstring]string from string, the values become key, and key values
func MakeReverseStringMap(text string, valueSepartor string, itemSeparator string) map[string]string {
	var result = make(map[string]string)
	for _, item := range strings.Split(text, itemSeparator) {
		if len(item) == 0 {
			continue
		}
		keyValue := strings.SplitN(item, valueSepartor, 2)
		if len(keyValue) == 2 {
			result[strings.Trim(keyValue[1], " \t")] = strings.Trim(keyValue[0], " \n\t")
		}
	}
	return result
}
