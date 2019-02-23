package data

import (
	"fmt"
	"github.com/viant/toolbox"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

func ExtractPath(expression string) string {
	var result = ""
	for _, r := range expression {
		aChar := string(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) || aChar == "[" || aChar == "]" || aChar == "." || aChar == "_" || aChar == "{" || aChar == "}" {
			result += aChar
		}
	}
	return strings.Trim(result, "{}")
}

func recordToMap(fields []*Field, record []interface{}, aMap map[string]interface{}) {
	for _, field := range fields {
		index := field.index
		var value = record[index]
		if value == nil {
			continue
		}
		aMap[field.Name] = value
	}
}

func indexValue(indexBy []int, record []interface{}) interface{} {
	if len(indexBy) == 1 {
		return record[indexBy[0]]
	}
	var values = make([]string, len(indexBy))
	for i, fieldIndex := range indexBy {
		values[i] = toolbox.AsString(record[fieldIndex])
	}
	return strings.Join(values, "-")
}

func intsToGenericSlice(keyType reflect.Type, aSlice []int) []interface{} {
	var result = make([]interface{}, len(aSlice))
	for i, item := range aSlice {
		result[i] = reflect.ValueOf(item).Convert(keyType).Interface()
	}
	return result
}

func floatsToGenericSlice(keyType reflect.Type, aSlice []float64) []interface{} {
	var result = make([]interface{}, len(aSlice))
	for i, item := range aSlice {
		result[i] = reflect.ValueOf(item).Convert(keyType).Interface()
	}
	return result
}

func stringsToGenericSlice(aSlice []string) []interface{} {
	var result = make([]interface{}, len(aSlice))
	for i, item := range aSlice {
		result[i] = toolbox.AsString(item)
	}
	return result
}

func sortKeys(key interface{}, aMap map[interface{}][]interface{}) ([]interface{}, error) {
	if len(aMap) == 0 {
		return []interface{}{}, nil
	}
	var i = 0
	switch key.(type) {
	case int, uint, uint8, uint16, uint32, uint64, int8, int16, int32, int64:
		var aSlice = make([]int, len(aMap))
		for k := range aMap {
			aSlice[i] = toolbox.AsInt(k)
			i++
		}
		sort.Ints(aSlice)
		return intsToGenericSlice(reflect.TypeOf(key), aSlice), nil
	case float64, float32:
		var aSlice = make([]float64, len(aMap))
		for k := range aMap {
			aSlice[i] = toolbox.AsFloat(k)
			i++
		}
		sort.Float64s(aSlice)
		return floatsToGenericSlice(reflect.TypeOf(key), aSlice), nil
	case string:
		var aSlice = make([]string, len(aMap))
		for k := range aMap {
			aSlice[i] = toolbox.AsString(k)
			i++
		}
		sort.Strings(aSlice)
		return stringsToGenericSlice(aSlice), nil
	}
	return nil, fmt.Errorf("unable sort, unsupported type: %T", key)
}
