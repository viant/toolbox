package toolbox

import (
	"fmt"
	"github.com/pkg/errors"
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

//ReverseSlice reverse a slice
func ReverseSlice(source interface{}) {
	if source == nil {
		return
	}
	var j = 0
	switch slice := source.(type) {
	case []byte:
		var sliceLen = len(slice)
		if sliceLen <= 1 {
			return
		}
		for i := sliceLen - 1; i >= (sliceLen / 2); i-- {
			item := slice[i]
			slice[i] = slice[j]
			slice[j] = item
			j++
		}
		return
	case []interface{}:
		var sliceLen = len(slice)
		if sliceLen <= 1 {
			return
		}
		for i := sliceLen - 1; i >= (sliceLen / 2); i-- {
			item := slice[i]
			slice[i] = slice[j]
			slice[j] = item
			j++
		}
		return
	case []string:
		var sliceLen = len(slice)
		if sliceLen <= 1 {
			return
		}
		for i := sliceLen - 1; i >= (sliceLen / 2); i-- {
			item := slice[i]
			slice[i] = slice[j]
			slice[j] = item
			j++
		}
		return
	}
	sliceValue := reflect.ValueOf(source)
	if sliceValue.IsNil() || !sliceValue.IsValid() {
		return
	}
	if sliceValue.Kind() == reflect.Ptr {
		sliceValue = sliceValue.Elem()
	}
	var sliceLen = sliceValue.Len()
	if sliceLen <= 1 {
		return
	}

	for i := sliceLen - 1; i >= (sliceLen / 2); i-- {
		indexItem := sliceValue.Index(i)
		indexItemValue := indexItem.Elem()
		if indexItem.Kind() == reflect.Ptr {
			sliceValue.Index(i).Set(sliceValue.Index(j).Elem().Addr())
			sliceValue.Index(j).Set(indexItemValue.Addr())
		} else {
			sliceValue.Index(i).Set(sliceValue.Index(j).Elem())
			sliceValue.Index(j).Set(indexItemValue)
		}
		j++
	}
}

//ProcessSlice iterates over any slice, it calls handler with each element unless handler returns false,
func ProcessSlice(slice interface{}, handler func(item interface{}) bool) {
	//The common cases with reflection for speed
	if aSlice, ok := slice.([]interface{}); ok {
		for _, item := range aSlice {
			if !handler(item) {
				break
			}

		}
		return
	}
	if aSlice, ok := slice.([]map[string]interface{}); ok {
		for _, item := range aSlice {
			if !handler(item) {
				break
			}

		}
		return
	}
	//The common cases with reflection for speed
	if aSlice, ok := slice.([]string); ok {
		for _, item := range aSlice {
			if !handler(item) {
				break
			}
		}
		return
	}

	//The common cases with reflection for speed
	if aSlice, ok := slice.([]int); ok {
		for _, item := range aSlice {
			if !handler(item) {
				break
			}
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
			if !handler(i, item) {
				break
			}
		}
		return
	}
	if aSlice, ok := slice.([]string); ok {
		for i, item := range aSlice {
			if !handler(i, item) {
				break
			}
		}
		return
	}
	if aSlice, ok := slice.([]int); ok {
		for i, item := range aSlice {
			if !handler(i, item) {
				break
			}
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

//AsSlice converts underlying slice or Ranger as []interface{}
func AsSlice(sourceSlice interface{}) []interface{} {
	var result []interface{}
	ranger, ok := sourceSlice.(Ranger)
	if ok {
		result = []interface{}{}
		_ = ranger.Range(func(item interface{}) (bool, error) {
			result = append(result, item)
			return true, nil
		})
		return result
	}
	iterator, ok := sourceSlice.(Iterator)
	if ok {
		if iterator.HasNext() {
			var item interface{}
			if err := iterator.Next(&item); err == nil {
				result = append(result, item)
			}
		}
	}
	result, ok = sourceSlice.([]interface{})
	if ok {
		return result
	}
	if resultPointer, ok := sourceSlice.(*[]interface{}); ok {
		return *resultPointer
	}
	result = make([]interface{}, 0)
	CopySliceElements(sourceSlice, &result)
	return result
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
	//optimized case
	if stringBoolMap, ok := targetMap.(map[string]bool); ok {
		if stringSlice, ok := sourceSlice.([]string); ok {
			if valueFunction, ok := keyFunction.(func(string) bool); ok {
				if keyFunction, ok := keyFunction.(func(string) string); ok {
					for _, item := range stringSlice {
						stringBoolMap[keyFunction(item)] = valueFunction(item)
					}
					return
				}
			}
		}
	}

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

var errSliceDoesNotHoldKeyValuePairs = errors.New("unable process map, not key value pairs")

//ProcessMap iterates over any map, it calls handler with every key, value pair unless handler returns false.
func ProcessMap(source interface{}, handler func(key, value interface{}) bool) error {
	switch aSlice := source.(type) {
	case map[string]string:
		for key, value := range aSlice {
			if !handler(key, value) {
				break
			}
		}
		return nil
	case map[string]interface{}:
		for key, value := range aSlice {
			if !handler(key, value) {
				break
			}
		}
		return nil
	case map[string]bool:
		for key, value := range aSlice {
			if !handler(key, value) {
				break
			}
		}
		return nil
	case map[string]int:
		for key, value := range aSlice {
			if !handler(key, value) {
				break
			}
		}
		return nil
	case map[interface{}]interface{}:
		for key, value := range aSlice {
			if !handler(key, value) {
				break
			}
		}
		return nil
	case map[int]interface{}:
		for key, value := range aSlice {
			if !handler(key, value) {
				break
			}
		}
		return nil

	}
	if IsSlice(source) {
		var err error
		var entryMap map[string]interface{}

		ProcessSlice(source, func(item interface{}) bool {
			entryMap, err = ToMap(item)
			if err != nil {
				return false
			}
			var key, value interface{}
			key, value, err = entryMapToKeyValue(entryMap)
			if err != nil {
				return false
			}
			return handler(key, value)
		})
		if err != nil {
			return errSliceDoesNotHoldKeyValuePairs
		}
		return nil
	}

	if !IsMap(source) {
		return errSliceDoesNotHoldKeyValuePairs
	}
	mapValue := DiscoverValueByKind(reflect.ValueOf(source), reflect.Map)
	for _, key := range mapValue.MapKeys() {
		value := mapValue.MapIndex(key)
		if !handler(key.Interface(), value.Interface()) {
			break
		}
	}
	return nil
}

//ToMap converts underlying map/struct/[]KV as map[string]interface{}
func ToMap(source interface{}) (map[string]interface{}, error) {
	if source == nil {
		return nil, nil
	}
	var result map[string]interface{}
	switch candidate := source.(type) {
	case map[string]interface{}:
		return candidate, nil
	case *map[string]interface{}:
		return *candidate, nil
	case map[interface{}]interface{}:
		result = make(map[string]interface{})
		for k, v := range candidate {
			result[AsString(k)] = v
		}
		return result, nil
	}
	if IsStruct(source) {
		var result = make(map[string]interface{})
		if err := DefaultConverter.AssignConverted(&result, source); err != nil {
			return nil, err
		}

		return result, nil
	} else if IsSlice(source) {
		var result = make(map[string]interface{})
		if err := DefaultConverter.AssignConverted(&result, source); err != nil {
			return nil, err
		}
		return result, nil
	}
	sourceMapValue := reflect.ValueOf(source)
	mapType := reflect.TypeOf(result)
	if sourceMapValue.Type().AssignableTo(mapType) {
		result, ok := sourceMapValue.Convert(mapType).Interface().(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unable to convert: %T to %T", source, map[string]interface{}{})
		}
		return result, nil
	}
	result = make(map[string]interface{})
	CopyMapEntries(source, result)
	return result, nil
}

//AsMap converts underlying map as map[string]interface{}
func AsMap(source interface{}) map[string]interface{} {
	if result, err := ToMap(source); err == nil {
		return result
	}
	return nil
}

//CopyMapEntries appends map entry from source map to target map
func CopyMapEntries(sourceMap, targetMap interface{}) {
	targetMapValue := reflect.ValueOf(targetMap)
	if targetMapValue.Kind() == reflect.Ptr {
		targetMapValue = targetMapValue.Elem()
	}
	if target, ok := targetMap.(map[string]interface{}); ok {
		_ = ProcessMap(sourceMap, func(key, value interface{}) bool {
			target[AsString(key)] = value
			return true
		})
		return
	}
	_ = ProcessMap(sourceMap, func(key, value interface{}) bool {
		targetMapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		return true
	})
}

//MapKeysToSlice appends all map keys to targetSlice
func MapKeysToSlice(sourceMap interface{}, targetSlicePointer interface{}) error {
	AssertPointerKind(targetSlicePointer, reflect.Slice, "targetSlicePointer")
	slicePointerValue := reflect.ValueOf(targetSlicePointer).Elem()
	return ProcessMap(sourceMap, func(key, value interface{}) bool {
		slicePointerValue.Set(reflect.Append(slicePointerValue, reflect.ValueOf(key)))
		return true
	})
}

//MapKeysToStringSlice creates a string slice from sourceMap keys, keys do not need to be of a string type.
func MapKeysToStringSlice(sourceMap interface{}) []string {
	var keys = make([]string, 0)
	//common cases
	switch aMap := sourceMap.(type) {
	case map[string]interface{}:
		for k := range aMap {
			keys = append(keys, k)
		}
		return keys
	case map[string]bool:
		for k := range aMap {
			keys = append(keys, k)
		}
		return keys
	case map[string]int:
		for k := range aMap {
			keys = append(keys, k)
		}
		return keys
	}
	_ = ProcessMap(sourceMap, func(key interface{}, value interface{}) bool {
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
func MakeStringMap(text string, valueSeparator string, itemSeparator string) map[string]string {
	var result = make(map[string]string)
	for _, item := range strings.Split(text, itemSeparator) {
		if len(item) == 0 {
			continue
		}
		keyValue := strings.SplitN(item, valueSeparator, 2)
		if len(keyValue) == 2 {
			result[strings.Trim(keyValue[0], " \t")] = strings.Trim(keyValue[1], " \n\t")
		}
	}
	return result
}

//MakeMap creates a mapstring]interface{} from string,
func MakeMap(text string, valueSeparator string, itemSeparator string) map[string]interface{} {
	var result = make(map[string]interface{})
	for _, item := range strings.Split(text, itemSeparator) {
		if len(item) == 0 {
			continue
		}
		keyValue := strings.SplitN(item, valueSeparator, 2)
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

func isNilOrEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch value := v.(type) {
	case string:
		if value == "" {
			return true
		}
	case int:
		if value == 0 {
			return true
		}
	case map[string]interface{}:
		if len(value) == 0 {
			return true
		}
	case map[interface{}]interface{}:
		if len(value) == 0 {
			return true
		}

	case []map[string]interface{}:
		if len(value) == 0 {
			return true
		}
	case []map[interface{}]interface{}:
		if len(value) == 0 {
			return true
		}
	case []interface{}:
		if len(value) == 0 {
			return true
		}
	case interface{}:
		if value == nil {
			return true
		}
	}
	return AsString(v) == ""
}




//CopyMap copy source map into destination map, copier function can modify key, value, or return false to skip map entry
func CopyMap(input, output interface{}, copier func(key, value interface{}) (interface{},interface{}, bool)) (err error) {
	var mutator func(k, v interface{})
	if aMap, ok := output.(map[interface{}]interface{}); ok {
		mutator = func(k, v interface{}) {
			aMap[k] = v
		}
	} else if aMap, ok := output.(map[string]interface{}); ok {
		mutator = func(k, v interface{}) {
			aMap[AsString(k)] = v
		}
	} else {
		return fmt.Errorf("unsupported map type: %v", output)
	}

	mapProvider := func(source interface{}) func() interface{} {
		if _, ok := source.(map[interface{}]interface{}); ok {
			return func() interface{} {
				return map[interface{}]interface{}{}
			}
		}
		return func() interface{} {
			return map[string]interface{}{}
		}
	}
	var ok bool
	err = ProcessMap(input, func(k, v interface{}) bool {
		k,  v,  ok = copier(k, v)
		if ! ok {
			return true
		}
		if v == nil {
			//
		} else if  IsMap(v) {
			transformed := mapProvider(v)()
			err = CopyMap(v, transformed, copier)
			if err != nil {
				return false
			}
			if isNilOrEmpty(transformed) {
				return true
			}
			v = transformed

		} else if  IsSlice(v) {
			aSlice := AsSlice(v)
			var transformed = []interface{}{}
			for _, item := range aSlice {
				if isNilOrEmpty(item) {
					continue
				}
				if IsMap(item) {
					transformedItem := mapProvider(item)()

					err = CopyMap(item, transformedItem, copier)
					if err != nil {
						return false
					}
					if isNilOrEmpty(transformedItem) {
						return true
					}
					transformed = append(transformed, transformedItem)
				} else {
					transformed = append(transformed, item)
				}

			}
			if len(transformed) == 0 {
				return true
			}
			v = transformed
		}
		mutator(k, v)
		return true
	})
	return err
}


//OmitEmptyMapWriter return false for all nil or empty values
func OmitEmptyMapWriter(key, value interface{}) (interface{}, interface{}, bool){
	if value == nil {
		return key, value, false
	}
	if IsPointer(value) {
		if reflect.ValueOf(value).IsNil() {
			return key, value, false
		}
	}
	if IsString(value) {
		return key, value,  AsString(value) != ""
	}
	if IsBool(value) {
		return key, value, AsBoolean(value)
	}
	if IsNumber(value) {
		return key, value, AsFloat(value) != 0.0
	}
	return key, value, true
}



//CopyNonEmptyMapEntries removes empty keys from map result
func CopyNonEmptyMapEntries(input, output interface{}) (err error) {
	return CopyMap(input, output, func(key, value interface{}) (interface{}, interface{}, bool) {
		if isNilOrEmpty(value) {
			return key, value, false
		}
		return key, value, true
	})
}

//CopyNonEmptyMapEntries removes empty keys from map result
func ReplaceMapEntries(input, output interface{}, replacement map[string]interface{}, removeEmpty bool) (err error) {
	return CopyMap(input, output, func(key, value interface{}) (interface{}, interface{}, bool) {
		k := AsString(key)
		if v, ok := replacement[k]; ok {
			return key, v, ok
		}
		if removeEmpty && isNilOrEmpty(value) {
			return nil, nil, false
		}
		return key, value, true
	})
}

//DeleteEmptyKeys removes empty keys from map result
func DeleteEmptyKeys(input interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	err := CopyNonEmptyMapEntries(input, result)
	if err == nil {
		return result
	}
	return AsMap(input)
}

//DeleteEmptyKeys removes empty keys from map result
func ReplaceMapKeys(input interface{}, replacement map[string]interface{}, removeEmpty bool) map[string]interface{} {
	result := map[string]interface{}{}
	_ = ReplaceMapEntries(input, result, replacement, removeEmpty)
	return result
}

//Pairs returns map for pairs.
func Pairs(params ...interface{}) map[string]interface{} {
	var result = make(map[string]interface{})
	for i := 0; i+1 < len(params); i += 2 {
		var key = AsString(params[i])
		result[key] = params[i+1]
	}
	return result
}

// Intersect find elements presents in slice a and b,  match is appended to result slice
//It accept generic slices,   All slices should have items of the same type or interface{} type
func Intersect(a, b interface{}, resultPointer interface{}) error {
	if reflect.ValueOf(resultPointer).Kind() != reflect.Ptr {
		return fmt.Errorf("resultPointer has to be pointer but had: %T", resultPointer)
	}
	var aItems = make(map[interface{}]bool)
	ProcessSlice(a, func(item interface{}) bool {
		aItems[item] = true
		return true
	})

	var appendMatch func(item interface{}) error
	switch aSlicePrt := resultPointer.(type) {
	case *[]interface{}:
		appendMatch = func(item interface{}) error {
			*aSlicePrt = append(*aSlicePrt, item)
			return nil
		}
	case *[]string:
		appendMatch = func(item interface{}) error {
			*aSlicePrt = append(*aSlicePrt, AsString(item))
			return nil
		}
	case *[]int:
		appendMatch = func(item interface{}) error {
			intValue, err := ToInt(item)
			if err != nil {
				return err
			}
			*aSlicePrt = append(*aSlicePrt, intValue)
			return nil
		}
	default:
		appendMatch = func(item interface{}) error {
			sliceValue := reflect.ValueOf(resultPointer).Elem()
			//TODO add check that type of the slice is assignable from item
			sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(item)))
			return nil
		}
	}
	var err error
	ProcessSlice(b, func(item interface{}) bool {
		if aItems[item] {
			if err = appendMatch(item); err != nil {
				return false
			}
		}
		return true
	})
	return err
}
