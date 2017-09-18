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
	switch sourceValue := input.(type) {
	case string:
		return sourceValue
	case []byte:
		return string(sourceValue)
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

func (c *Converter) assignConvertedSlice(target, source interface{}, targetIndirectValue reflect.Value, targetIndirectPointerType reflect.Type) error {
	sliceType := DiscoverTypeByKind(target, reflect.Slice)
	slicePointer := reflect.New(sliceType)
	slice := slicePointer.Elem()

	componentType := DiscoverComponentType(target)
	var err error
	ProcessSlice(source, func(item interface{}) bool {
		var targetComponentPointer = reflect.New(componentType)

		if componentType.Kind() == reflect.Map {
			targetComponent := reflect.MakeMap(componentType)
			targetComponentPointer.Elem().Set(targetComponent)
		}
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
		mapping, found := fieldsMapping[strings.ToLower(key)]
		if found {

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

//AssignConverted assign to the target source, target needs to be pointer, input has to be convertible or compatible type
func (c *Converter) AssignConverted(target, source interface{}) error {
	if target == nil {
		return fmt.Errorf("destinationPointer was nil %v %v", target, source)
	}
	if source == nil {
		return nil
	}

	switch targetValuePointer := target.(type) {
	case *string:
		switch sourceValue := source.(type) {
		case string:
			*targetValuePointer = sourceValue
			return nil
		case *string:
			*targetValuePointer = *sourceValue
			return nil
		case []byte:
			*targetValuePointer = string(sourceValue)
			return nil
		case *[]byte:
			*targetValuePointer = string(NewBytes(*sourceValue))
			return nil
		default:
			*targetValuePointer = AsString(source)
			return nil
		}

	case **string:
		switch sourceValue := source.(type) {
		case string:
			*targetValuePointer = &sourceValue
			return nil
		case *string:
			*targetValuePointer = sourceValue
			return nil
		case []byte:
			var stringSourceValue = string(sourceValue)
			*targetValuePointer = &stringSourceValue
			return nil
		case *[]byte:
			var stringSourceValue = string(NewBytes(*sourceValue))
			*targetValuePointer = &stringSourceValue
			return nil
		default:
			stringSourceValue := AsString(source)
			*targetValuePointer = &stringSourceValue
			return nil
		}

	case *[]string:
		switch sourceValue := source.(type) {
		case []string:
			*targetValuePointer = sourceValue
			return nil
		case *[]string:
			*targetValuePointer = *sourceValue
			return nil
		default:
			var stingItems = make([]string, 0)
			ProcessSlice(source, func(item interface{}) bool {
				stingItems = append(stingItems, AsString(item))
				return true
			})
			*targetValuePointer = stingItems
			return nil
		}

	case *bool:
		switch sourceValue := source.(type) {
		case bool:
			*targetValuePointer = sourceValue
			return nil
		case *bool:
			*targetValuePointer = *sourceValue
			return nil

		case int:
			*targetValuePointer = sourceValue != 0
			return nil
		case string:
			boolValue, err := strconv.ParseBool(sourceValue)
			if err != nil {
				return err
			}

			*targetValuePointer = boolValue
			return nil
		case *string:
			boolValue, err := strconv.ParseBool(*sourceValue)
			if err != nil {
				return err
			}
			*targetValuePointer = boolValue
			return nil
		}

	case **bool:
		switch sourceValue := source.(type) {
		case bool:
			*targetValuePointer = &sourceValue
			return nil
		case *bool:
			*targetValuePointer = sourceValue
			return nil
		case int:
			boolValue := sourceValue != 0
			*targetValuePointer = &boolValue
			return nil
		case string:
			boolValue, err := strconv.ParseBool(sourceValue)
			if err != nil {
				return err
			}

			*targetValuePointer = &boolValue
			return nil
		case *string:
			boolValue, err := strconv.ParseBool(*sourceValue)
			if err != nil {
				return err
			}
			*targetValuePointer = &boolValue
			return nil
		}
	case *[]byte:
		switch sourceValue := source.(type) {
		case []byte:
			*targetValuePointer = sourceValue
			return nil
		case *[]byte:
			*targetValuePointer = *sourceValue
			return nil
		case string:
			*targetValuePointer = []byte(sourceValue)
			return nil
		case *string:
			var stringValue = *sourceValue
			*targetValuePointer = []byte(stringValue)
			return nil
		}

	case **[]byte:
		switch sourceValue := source.(type) {
		case []byte:
			bytes := NewBytes(sourceValue)
			*targetValuePointer = &bytes
			return nil
		case *[]byte:
			bytes := NewBytes(*sourceValue)
			*targetValuePointer = &bytes
			return nil
		case string:
			bytes := []byte(sourceValue)
			*targetValuePointer = &bytes
			return nil
		case *string:
			bytes := []byte(*sourceValue)
			*targetValuePointer = &bytes
			return nil
		}

	case *int, *int8, *int16, *int32, *int64:
		directValue := reflect.Indirect(reflect.ValueOf(targetValuePointer))
		sourceValue := reflect.ValueOf(source)
		if sourceValue.Kind() == reflect.Ptr {
			sourceValue = sourceValue.Elem()
		}
		stringValue := AsString(sourceValue.Interface())
		value, err := strconv.ParseInt(stringValue, 10, directValue.Type().Bits())
		if err != nil {
			return err
		}
		directValue.SetInt(value)
		return nil

	case **int, **int8, **int16, **int32, **int64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		sourceValue := reflect.ValueOf(source)
		if sourceValue.Kind() == reflect.Ptr {
			sourceValue = sourceValue.Elem()
		}
		stringValue := AsString(sourceValue.Interface())

		value, err := strconv.ParseInt(stringValue, 10, directType.Bits())
		if err != nil {
			return err
		}
		reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&value))
		return nil
	case *uint, *uint8, *uint16, *uint32, *uint64:
		directValue := reflect.Indirect(reflect.ValueOf(targetValuePointer))
		sourceValue := reflect.ValueOf(source)
		if sourceValue.Kind() == reflect.Ptr {
			sourceValue = sourceValue.Elem()
		}
		stringValue := AsString(sourceValue.Interface())
		value, err := strconv.ParseUint(stringValue, 10, directValue.Type().Bits())
		if err != nil {
			return err
		}
		directValue.SetUint(value)
		return nil
	case **uint, **uint8, **uint16, **uint32, **uint64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		sourceValue := reflect.ValueOf(source)
		if sourceValue.Kind() == reflect.Ptr {
			sourceValue = sourceValue.Elem()
		}
		stringValue := AsString(sourceValue.Interface())

		value, err := strconv.ParseUint(stringValue, 10, directType.Bits())
		if err != nil {
			return err
		}
		reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&value))
		return nil
	case *float32, *float64:
		directValue := reflect.Indirect(reflect.ValueOf(targetValuePointer))
		sourceValue := reflect.ValueOf(source)
		if sourceValue.Kind() == reflect.Ptr {
			sourceValue = sourceValue.Elem()
		}
		stringValue := AsString(sourceValue.Interface())
		value, err := strconv.ParseFloat(stringValue, directValue.Type().Bits())
		if err != nil {
			return err
		}
		directValue.SetFloat(value)
		return nil
	case **float32, **float64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		sourceValue := reflect.ValueOf(source)
		if sourceValue.Kind() == reflect.Ptr {
			sourceValue = sourceValue.Elem()
		}
		stringValue := AsString(sourceValue.Interface())
		value, err := strconv.ParseFloat(stringValue, directType.Bits())
		if err != nil {
			return err
		}
		reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&value))
		return nil
	case *time.Time:
		switch sourceValue := source.(type) {
		case string:
			timeValue := AsTime(sourceValue, c.DataLayout)
			if timeValue == nil {
				_, err := time.Parse(c.DataLayout, sourceValue)
				return err
			}
			*targetValuePointer = *timeValue
			return nil
		case *string:
			timeValue := AsTime(sourceValue, c.DataLayout)
			if timeValue == nil {
				_, err := time.Parse(c.DataLayout, *sourceValue)
				return err
			}
			*targetValuePointer = *timeValue
			return nil
		case int, int64, uint, uint64, float32, float64, *int, *int64, *uint, *uint64, *float32, *float64:
			intValue := int(AsFloat(sourceValue))
			timeValue := time.Unix(int64(intValue), 0)
			*targetValuePointer = timeValue
			return nil

		}

	case **time.Time:
		switch sourceValue := source.(type) {
		case string:
			timeValue := AsTime(sourceValue, c.DataLayout)
			if timeValue == nil {
				_, err := time.Parse(c.DataLayout, sourceValue)
				return err
			}
			*targetValuePointer = timeValue
			return nil
		case *string:
			timeValue := AsTime(sourceValue, c.DataLayout)
			if timeValue == nil {
				_, err := time.Parse(c.DataLayout, *sourceValue)
				return err
			}
			*targetValuePointer = timeValue
			return nil
		case int, int64, uint, uint64, float32, float64, *int, *int64, *uint, *uint64, *float32, *float64:
			intValue := int(AsFloat(sourceValue))
			timeValue := time.Unix(int64(intValue), 0)
			*targetValuePointer = &timeValue
			return nil

		}

	case *interface{}:

		(*targetValuePointer) = source
		return nil
	case **interface{}:
		(*targetValuePointer) = &source
		return nil

	}

	sourceValue := reflect.ValueOf(source)
	if source == nil || !sourceValue.IsValid() || (sourceValue.CanSet() && sourceValue.IsNil()) {
		return nil
	}

	targetIndirectValue := reflect.Indirect(reflect.ValueOf(target))
	if sourceValue.IsValid() && sourceValue.Type().AssignableTo(reflect.TypeOf(target)) {
		targetIndirectValue.Set(sourceValue.Elem())
		return nil
	}

	var targetIndirectPointerType = reflect.TypeOf(target).Elem()
	if targetIndirectPointerType.Kind() == reflect.Ptr || targetIndirectPointerType.Kind() == reflect.Slice || targetIndirectPointerType.Kind() == reflect.Map {
		targetIndirectPointerType = targetIndirectPointerType.Elem()

	}

	if targetIndirectValue.Kind() == reflect.Slice || targetIndirectPointerType.Kind() == reflect.Slice {

		if sourceValue.Kind() == reflect.Ptr && sourceValue.Elem().Kind() == reflect.Slice {
			sourceValue = sourceValue.Elem()
		}
		if sourceValue.Kind() == reflect.Slice {
			return c.assignConvertedSlice(target, source, targetIndirectValue, targetIndirectPointerType)
		}
	}
	if targetIndirectValue.Kind() == reflect.Map || targetIndirectPointerType.Kind() == reflect.Map {

		if sourceValue.Kind() == reflect.Ptr {
			sourceValue = sourceValue.Elem()
		}

		if sourceValue.Kind() == reflect.Map {
			return c.assignConvertedMap(target, source, targetIndirectValue, targetIndirectPointerType)
		} else if sourceValue.Kind() == reflect.Struct {
			return c.assignConvertedMapFromStruct(source, target, sourceValue)
		}

	} else if targetIndirectValue.Kind() == reflect.Struct {

		inputMap := AsMap(source)
		if inputMap != nil {
			err := c.assignConvertedStruct(target, inputMap, targetIndirectValue, targetIndirectPointerType)
			return err
		}

	} else if targetIndirectPointerType.Kind() == reflect.Struct {

		structPointer := reflect.New(targetIndirectPointerType)
		inputMap := AsMap(source)
		if inputMap != nil {
			err := c.assignConvertedStruct(target, inputMap, structPointer.Elem(), targetIndirectPointerType)
			if err != nil {
				return err
			}
			targetIndirectValue.Set(structPointer)
			return nil
		}

	}

	if sourceValue.IsValid() && sourceValue.Type().AssignableTo(targetIndirectValue.Type()) {
		targetIndirectValue.Set(sourceValue)
		return nil
	}
	if sourceValue.IsValid() && sourceValue.Type().ConvertibleTo(targetIndirectValue.Type()) {
		converted := sourceValue.Convert(targetIndirectValue.Type())
		targetIndirectValue.Set(converted)
		return nil
	}

	return fmt.Errorf("Unable to convert type %T into type %T", source, target)
}

func (c *Converter) assignConvertedMapFromStruct(source, target interface{}, sourceValue reflect.Value) error {
	targetMap := AsMap(target)
	if targetMap == nil {
		return fmt.Errorf("target %T is not a map", target)
	}
	sourceType := sourceValue.Type()
	for i := 0; i < sourceValue.NumField(); i++ {
		field := sourceValue.Field(i)
		if !field.CanAddr() {
			continue
		}
		var value interface{}
		fieldInfo := field
		fieldKind := field.Kind()

		for fieldKind == reflect.Ptr {
			fieldInfo = fieldInfo.Elem()
			fieldKind = fieldInfo.Kind()
		}
		fieldType := sourceType.Field(i)
		if fieldKind == reflect.Struct {
			aMap := make(map[string]interface{})
			err := c.AssignConverted(&aMap, field.Interface())
			if err != nil {
				return err
			}
			value = aMap
		} else if fieldKind == reflect.Slice {

			var componentType = DiscoverComponentType(field.Interface())
			for componentType.Kind() == reflect.Ptr {
				componentType = componentType.Elem()
			}
			if componentType.Kind() == reflect.Struct {
				slice := make([]map[string]interface{}, 0)
				err := c.AssignConverted(&slice, field.Interface())
				if err != nil {
					return err
				}
				value = slice
			} else {
				slice := make([]interface{}, 0)
				err := c.AssignConverted(&slice, field.Interface())
				if err != nil {
					return err
				}
				value = slice
			}

		} else {
			err := c.AssignConverted(&value, field.Interface())
			if err != nil {
				return err
			}
		}
		targetMap[fieldType.Name] = value

	}
	return nil
}

//NewColumnConverter create a new converter, that has abbility to convert map to struct using column mapping
func NewColumnConverter(dataFormat string) *Converter {
	return &Converter{dataFormat, "column"}
}
