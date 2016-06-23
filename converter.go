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

// Package toolbox - data type conversion, data converter
package toolbox

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//DefaultDateLayout is set to 2006-01-02 15:04:05.000
var DefaultDateLayout = "2006-01-02 15:04:05.000"

//AsString converts an input to string.
func AsString(input interface{}) string {
	switch inputValue := input.(type) {
	case string:
		return inputValue
	case []byte:
		return string(inputValue)
	}

	reflectValue := reflect.ValueOf(input)
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	switch reflectValue.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(reflectValue.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(reflectValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(reflectValue.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(reflectValue.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(reflectValue.Float(), 'g', -1, 32)
	}
	return fmt.Sprintf("%v", input)
}

//CanConvertToFloat checkis if float conversion is possible.
func CanConvertToFloat(value interface{}) bool {
	if _, ok := value.(float64); ok {
		return true
	}
	_, err := strconv.ParseFloat(AsString(value), 64)
	return err == nil
}

//AsFloat converts an input to float.
func AsFloat(value interface{}) float64 {
	if floatValue, ok := value.(float64); ok {
		return floatValue
	}
	valueAsString := AsString(value)
	if result, err := strconv.ParseFloat(valueAsString, 64); err == nil {
		return result
	}
	return 0
}

//AsBoolean converts an input to bool.
func AsBoolean(value interface{}) bool {
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	valueAsString := AsString(value)
	if result, err := strconv.ParseBool(valueAsString); err == nil {
		return result
	}
	return false
}

//CanConvertToInt returns true if an input can be converted to int value.
func CanConvertToInt(value interface{}) bool {
	if _, ok := value.(int); ok {
		return true
	}
	valueAsString := AsString(value)
	if _, err := strconv.ParseInt(valueAsString, 10, 64); err == nil {
		return true
	}
	return false
}

//AsInt converts an input to int.
func AsInt(value interface{}) int {
	if intValue, ok := value.(int); ok {
		return intValue
	}
	if floatValue, ok := value.(float64); ok {
		return int(floatValue)
	}
	valueAsString := AsString(value)
	if result, err := strconv.ParseInt(valueAsString, 10, 64); err == nil {
		return int(result)
	}
	return 0
}

//AsTime converts an input to time, it takes time input,  dateLaout as parameters.
func AsTime(value interface{}, dateLayout string) *time.Time {
	if timeValue, ok := value.(time.Time); ok {
		return &timeValue
	}
	if CanConvertToFloat(value) {
		unixTime := int(AsFloat(value))
		timeValue := time.Unix(int64(unixTime), 0)
		return &timeValue
	}
	timeValue, err := ParseTime(AsString(value), dateLayout)
	if err != nil {
		return nil
	}
	return &timeValue
}

//DiscoverValueAndKind discovers input kind, it applies checks of the following types:  int, float, bool, string
func DiscoverValueAndKind(input string) (interface{}, reflect.Kind) {
	if len(input) == 0 {
		return nil, reflect.Invalid
	}
	if strings.Contains(input, ".") {
		if floatValue, err := strconv.ParseFloat(input, 64); err == nil {
			return floatValue, reflect.Float64
		}
	}
	if intValue, err := strconv.ParseInt(input, 10, 64); err == nil {
		return int(intValue), reflect.Int
	} else if strings.ToLower(input) == "true" {
		return true, reflect.Bool
	} else if strings.ToLower(input) == "false" {
		return false, reflect.Bool
	}
	return input, reflect.String
}

//DiscoverCollectionValuesAndKind discovers passed in slice item kind, and returns slice of values converted to discovered type.
//It tries the following kind int, float, bool, string
func DiscoverCollectionValuesAndKind(values interface{}) ([]interface{}, reflect.Kind) {
	var candidateKind = reflect.Int
	var result = make([]interface{}, 0)
	ProcessSlice(values, func(value interface{}) bool {
		stringValue := strings.ToLower(AsString(value))
		switch candidateKind {
		case reflect.String:
			return false
		case reflect.Int:
			if !strings.Contains(stringValue, ".") && CanConvertToInt(value) {
				return true
			}
			candidateKind = reflect.Float64
			fallthrough
		case reflect.Float64:
			if CanConvertToFloat(value) {
				return true
			}
			candidateKind = reflect.Bool
			fallthrough

		case reflect.Bool:
			if stringValue == "true" || stringValue == "false" {
				return true
			}
			candidateKind = reflect.String
		}
		return true
	})
	ProcessSlice(values, func(value interface{}) bool {
		switch candidateKind {
		case reflect.Float64:
			result = append(result, AsFloat(value))
		case reflect.Int:
			result = append(result, AsInt(value))
		case reflect.Bool:
			result = append(result, AsBoolean(value))
		default:
			result = append(result, AsString(value))
		}
		return true
	})
	return result, candidateKind
}

//UnwrapValue returns  value
func UnwrapValue(value *reflect.Value) interface{} {
	return value.Interface()
}

//NewBytes copies from input
func NewBytes(input []byte) []byte {
	if input != nil {
		var result = make([]byte, len(input))
		copy(result, input)
		return result
	}
	return nil
}

//ParseTime parses time, adjusting date layout to length of input
func ParseTime(input, layout string) (time.Time, error) {

	if len(layout) == 0 {
		layout = DefaultDateLayout
	} //GetFieldValue returns field value
	lastPosition := len(input)
	if lastPosition >= len(layout) {
		lastPosition = len(layout)
	}
	layout = layout[0:lastPosition]

	return time.Parse(layout, input)
}

//Converter represets data converter, it converts incompatibe data structure, like map and struct, string and time, *string to string, etc.
type Converter struct {
	DataLayout   string
	MappedKeyTag string
}

func (c *Converter) assignConvertedMap(target, input interface{}, targetIndirectValue reflect.Value, targetIndirectPointerType reflect.Type) error {
	mapType := DiscoverTypeByKind(target, reflect.Map)
	mapPointer := reflect.New(mapType)
	mapValueType := mapType.Elem()
	keyKeyType := mapType.Key()
	newMap := mapPointer.Elem()
	newMap.Set(reflect.MakeMap(mapType))
	var err error
	ProcessMap(input, func(key, value interface{}) bool {
		mapValueType = reflect.TypeOf(value)
		targetMapValuePointer := reflect.New(mapValueType)
		err = c.AssignConverted(targetMapValuePointer.Interface(), value)
		if err != nil {
			err = fmt.Errorf("Failed to assigned converted map value %v to %v due to %v", input, target, err)
			return false
		}

		targetMapKeyPointer := reflect.New(keyKeyType)
		err = c.AssignConverted(targetMapKeyPointer.Interface(), key)
		if err != nil {
			err = fmt.Errorf("Failed to assigned converted map key %v to %v due to %v", input, target, err)
			return false
		}
		newMap.SetMapIndex(targetMapKeyPointer.Elem(), targetMapValuePointer.Elem())
		return true
	})

	if targetIndirectPointerType.Kind() == reflect.Map {
		targetIndirectValue.Set(mapPointer)
	} else {
		targetIndirectValue.Set(newMap)
	}
	return err

}

func (c *Converter) assignConvertedSlice(target, input interface{}, targetIndirectValue reflect.Value, targetIndirectPointerType reflect.Type) error {
	sliceType := DiscoverTypeByKind(target, reflect.Slice)
	slicePointer := reflect.New(sliceType)
	slice := slicePointer.Elem()
	componentType := DiscoverComponentType(target)
	var err error
	ProcessSlice(input, func(item interface{}) bool {
		targetComponentPointer := reflect.New(componentType)
		err = c.AssignConverted(targetComponentPointer.Interface(), item)
		if err != nil {
			err = fmt.Errorf("Failed to convert slice tiem %v to %v due to %v", item, targetComponentPointer.Interface(), err)
			return false
		}
		slice.Set(reflect.Append(slice, targetComponentPointer.Elem()))
		return true
	})

	if targetIndirectPointerType.Kind() == reflect.Slice {
		targetIndirectValue.Set(slicePointer)
	} else {
		targetIndirectValue.Set(slice)
	}
	return err
}

func (c *Converter) assignConvertedStruct(target interface{}, inputMap map[string]interface{}, targetIndirectValue reflect.Value, targetIndirectPointerType reflect.Type) error {
	newStructPointer := reflect.New(targetIndirectValue.Type())
	newStruct := newStructPointer.Elem()
	fieldsMapping := NewFieldSettingByKey(newStructPointer.Interface(), c.MappedKeyTag)
	for key, value := range inputMap {
		if mapping, ok := fieldsMapping[strings.ToLower(key)]; ok {
			fieldName := mapping["fieldName"]
			field := newStruct.FieldByName(fieldName)

			if HasTimeLayout(mapping) {
				previousLayout := c.DataLayout
				c.DataLayout = GetTimeLayout(mapping)
				err := c.AssignConverted(field.Addr().Interface(), value)
				if err != nil {
					return fmt.Errorf("Failed to convert %v to %v due to %v", value, field, err)
				}
				c.DataLayout = previousLayout

			} else {
				err := c.AssignConverted(field.Addr().Interface(), value)
				if err != nil {
					return fmt.Errorf("Failed to convert %v to %v due to %v", value, field, err)
				}
			}
		}
	}

	if targetIndirectPointerType.Kind() == reflect.Slice {
		targetIndirectValue.Set(newStructPointer)
	} else {
		targetIndirectValue.Set(newStruct)
	}
	return nil
}

//AssignConverted assign to the target input, target needs to be pointer, input has to be convertible or compatible type
func (c *Converter) AssignConverted(target, input interface{}) error {
	if target == nil {
		return fmt.Errorf("destinationPointer was nil %v %v", target, input)
	}
	if input == nil {
		return nil
	}

	switch targetValuePointer := target.(type) {
	case *string:
		switch inputValue := input.(type) {
		case string:
			*targetValuePointer = inputValue
			return nil
		case *string:
			*targetValuePointer = *inputValue
			return nil
		case []byte:
			*targetValuePointer = string(inputValue)
			return nil
		case *[]byte:
			*targetValuePointer = string(NewBytes(*inputValue))
			return nil
		default:
			*targetValuePointer = AsString(input)
			return nil
		}

	case **string:
		switch inputValue := input.(type) {
		case string:
			*targetValuePointer = &inputValue
			return nil
		case *string:
			*targetValuePointer = inputValue
			return nil
		case []byte:
			var stringSourceValue = string(inputValue)
			*targetValuePointer = &stringSourceValue
			return nil
		case *[]byte:
			var stringSourceValue = string(NewBytes(*inputValue))
			*targetValuePointer = &stringSourceValue
			return nil
		default:
			stringSourceValue := AsString(input)
			*targetValuePointer = &stringSourceValue
			return nil
		}

	case *bool:
		switch inputValue := input.(type) {
		case bool:
			*targetValuePointer = inputValue
			return nil
		case *bool:
			*targetValuePointer = *inputValue
			return nil

		case int:
			*targetValuePointer = inputValue != 0
			return nil
		case string:
			boolValue, err := strconv.ParseBool(inputValue)
			if err != nil {
				return err
			}

			*targetValuePointer = boolValue
			return nil
		case *string:
			boolValue, err := strconv.ParseBool(*inputValue)
			if err != nil {
				return err
			}
			*targetValuePointer = boolValue
			return nil
		}

	case **bool:
		switch inputValue := input.(type) {
		case bool:
			*targetValuePointer = &inputValue
			return nil
		case *bool:
			*targetValuePointer = inputValue
			return nil
		case int:
			boolValue := inputValue != 0
			*targetValuePointer = &boolValue
			return nil
		case string:
			boolValue, err := strconv.ParseBool(inputValue)
			if err != nil {
				return err
			}

			*targetValuePointer = &boolValue
			return nil
		case *string:
			boolValue, err := strconv.ParseBool(*inputValue)
			if err != nil {
				return err
			}
			*targetValuePointer = &boolValue
			return nil
		}
	case *[]byte:
		switch inputValue := input.(type) {
		case []byte:
			*targetValuePointer = inputValue
			return nil
		case *[]byte:
			*targetValuePointer = *inputValue
			return nil
		case string:
			*targetValuePointer = []byte(inputValue)
			return nil
		case *string:
			var stringValue = *inputValue
			*targetValuePointer = []byte(stringValue)
			return nil
		}

	case **[]byte:
		switch inputValue := input.(type) {
		case []byte:
			bytes := NewBytes(inputValue)
			*targetValuePointer = &bytes
			return nil
		case *[]byte:
			bytes := NewBytes(*inputValue)
			*targetValuePointer = &bytes
			return nil
		case string:
			bytes := []byte(inputValue)
			*targetValuePointer = &bytes
			return nil
		case *string:
			bytes := []byte(*inputValue)
			*targetValuePointer = &bytes
			return nil
		}

	case *int, *int8, *int16, *int32, *int64:
		directValue := reflect.Indirect(reflect.ValueOf(targetValuePointer))
		inputValue := reflect.ValueOf(input)
		if inputValue.Kind() == reflect.Ptr {
			inputValue = inputValue.Elem()
		}
		stringValue := AsString(inputValue.Interface())
		value, err := strconv.ParseInt(stringValue, 10, directValue.Type().Bits())
		if err != nil {
			return err
		}
		directValue.SetInt(value)
		return nil

	case **int, **int8, **int16, **int32, **int64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		inputValue := reflect.ValueOf(input)
		if inputValue.Kind() == reflect.Ptr {
			inputValue = inputValue.Elem()
		}
		stringValue := AsString(inputValue.Interface())

		value, err := strconv.ParseInt(stringValue, 10, directType.Bits())
		if err != nil {
			return err
		}
		reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&value))
		return nil
	case *uint, *uint8, *uint16, *uint32, *uint64:
		directValue := reflect.Indirect(reflect.ValueOf(targetValuePointer))
		inputValue := reflect.ValueOf(input)
		if inputValue.Kind() == reflect.Ptr {
			inputValue = inputValue.Elem()
		}
		stringValue := AsString(inputValue.Interface())
		value, err := strconv.ParseUint(stringValue, 10, directValue.Type().Bits())
		if err != nil {
			return err
		}
		directValue.SetUint(value)
		return nil
	case **uint, **uint8, **uint16, **uint32, **uint64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		inputValue := reflect.ValueOf(input)
		if inputValue.Kind() == reflect.Ptr {
			inputValue = inputValue.Elem()
		}
		stringValue := AsString(inputValue.Interface())

		value, err := strconv.ParseUint(stringValue, 10, directType.Bits())
		if err != nil {
			return err
		}
		reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&value))
		return nil
	case *float32, *float64:
		directValue := reflect.Indirect(reflect.ValueOf(targetValuePointer))
		inputValue := reflect.ValueOf(input)
		if inputValue.Kind() == reflect.Ptr {
			inputValue = inputValue.Elem()
		}
		stringValue := AsString(inputValue.Interface())
		value, err := strconv.ParseFloat(stringValue, directValue.Type().Bits())
		if err != nil {
			return err
		}
		directValue.SetFloat(value)
		return nil
	case **float32, **float64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		inputValue := reflect.ValueOf(input)
		if inputValue.Kind() == reflect.Ptr {
			inputValue = inputValue.Elem()
		}
		stringValue := AsString(inputValue.Interface())
		value, err := strconv.ParseFloat(stringValue, directType.Bits())
		if err != nil {
			return err
		}
		reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&value))
		return nil
	case *time.Time:
		switch inputValue := input.(type) {
		case string:
			timeValue := AsTime(inputValue, c.DataLayout)
			if timeValue == nil {
				_, err := time.Parse(c.DataLayout, inputValue)
				return err
			}
			*targetValuePointer = *timeValue
			return nil
		case *string:
			timeValue := AsTime(inputValue, c.DataLayout)
			if timeValue == nil {
				_, err := time.Parse(c.DataLayout, *inputValue)
				return err
			}
			*targetValuePointer = *timeValue
			return nil
		case int, int64, uint, uint64, float32, float64, *int, *int64, *uint, *uint64, *float32, *float64:
			intValue := int(AsFloat(inputValue))
			timeValue := time.Unix(int64(intValue), 0)
			*targetValuePointer = timeValue
			return nil

		}

	case **time.Time:
		switch inputValue := input.(type) {
		case string:
			timeValue := AsTime(inputValue, c.DataLayout)
			if timeValue == nil {
				_, err := time.Parse(c.DataLayout, inputValue)
				return err
			}
			*targetValuePointer = timeValue
			return nil
		case *string:
			timeValue := AsTime(inputValue, c.DataLayout)
			if timeValue == nil {
				_, err := time.Parse(c.DataLayout, *inputValue)
				return err
			}
			*targetValuePointer = timeValue
			return nil
		case int, int64, uint, uint64, float32, float64, *int, *int64, *uint, *uint64, *float32, *float64:
			intValue := int(AsFloat(inputValue))
			timeValue := time.Unix(int64(intValue), 0)
			*targetValuePointer = &timeValue
			return nil

		}

	case *interface{}:
		(*targetValuePointer) = input
		return nil
	case **interface{}:
		(*targetValuePointer) = &input
		return nil

	}

	inputValue := reflect.ValueOf(input)
	if input == nil || !inputValue.IsValid() || (inputValue.CanSet() && inputValue.IsNil()) {
		return nil
	}

	targetIndirectValue := reflect.Indirect(reflect.ValueOf(target))
	if inputValue.IsValid() && inputValue.Type().AssignableTo(reflect.TypeOf(target)) {
		targetIndirectValue.Set(inputValue.Elem())
		return nil
	}

	var targetIndirectPointerType = reflect.TypeOf(target).Elem()
	if targetIndirectPointerType.Kind() == reflect.Ptr || targetIndirectPointerType.Kind() == reflect.Slice || targetIndirectPointerType.Kind() == reflect.Map {
		targetIndirectPointerType = targetIndirectPointerType.Elem()
	}

	if targetIndirectValue.Kind() == reflect.Slice || targetIndirectPointerType.Kind() == reflect.Slice {
		if inputValue.Kind() == reflect.Slice {
			return c.assignConvertedSlice(target, input, targetIndirectValue, targetIndirectPointerType)
		}
	}

	if targetIndirectValue.Kind() == reflect.Map || targetIndirectPointerType.Kind() == reflect.Map {
		if inputValue.Kind() == reflect.Map {
			return c.assignConvertedMap(target, input, targetIndirectValue, targetIndirectPointerType)
		}
	} else if targetIndirectValue.Kind() == reflect.Struct {
		if inputMap, ok := input.(map[string]interface{}); ok {
			return c.assignConvertedStruct(target, inputMap, targetIndirectValue, targetIndirectPointerType)
		}
	}

	if inputValue.IsValid() && inputValue.Type().AssignableTo(targetIndirectValue.Type()) {
		targetIndirectValue.Set(inputValue)
		return nil
	}
	if inputValue.IsValid() && inputValue.Type().ConvertibleTo(targetIndirectValue.Type()) {
		converted := inputValue.Convert(targetIndirectValue.Type())
		targetIndirectValue.Set(converted)
		return nil
	}
	return fmt.Errorf("Unable to convert type %T into type %T", input, target)
}

//NewColumnConverter create a new converter, that has abbility to convert map to struct using column mapping
func NewColumnConverter(dataFormat string) *Converter {
	return &Converter{dataFormat, "column"}
}
