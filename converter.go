package toolbox

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//DefaultDateLayout is set to 2006-01-02 15:04:05.000
var DefaultDateLayout = "2006-01-02 15:04:05.000"
var numericTypes = []reflect.Type{
	reflect.TypeOf(int(0)),
	reflect.TypeOf(int8(0)),
	reflect.TypeOf(int16(0)),
	reflect.TypeOf(int32(0)),
	reflect.TypeOf(int64(0)),

	reflect.TypeOf(uint(0)),
	reflect.TypeOf(uint8(0)),
	reflect.TypeOf(uint16(0)),
	reflect.TypeOf(uint32(0)),
	reflect.TypeOf(uint64(0)),

	reflect.TypeOf(float32(0.0)),
	reflect.TypeOf(float64(0.0)),
}

//AsString converts an input to string.
func AsString(input interface{}) string {
	switch value := input.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	case []interface{}:
		if len(value) == 0 {
			return ""
		}
		if _, isByte := value[0].(byte); isByte {
			var stringBytes = make([]byte, len(value))
			for i, v := range value {
				stringBytes[i] = v.(byte)
			}
			return string(stringBytes)
		}
		var result = ""
		for _, v := range value {
			result += AsString(v)
		}
		return result
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
	if result, err := ToFloat(value); err == nil {
		return result
	}
	return 0
}

//ToFloat converts an input to float or error
func ToFloat(value interface{}) (float64, error) {
	if value == nil {
		return 0, NewNilPointerError("float value was nil")
	}
	switch actualValue := value.(type) {
	case float64:
		return actualValue, nil
	case int:
		return float64(actualValue), nil
	case uint:
		return float64(actualValue), nil
	case int64:
		return float64(actualValue), nil
	case uint64:
		return float64(actualValue), nil
	case int32:
		return float64(actualValue), nil
	case uint32:
		return float64(actualValue), nil
	case float32:
		return float64(actualValue), nil
	case bool:
		if actualValue {
			return 1.0, nil
		}
		return 0.0, nil
	}
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		floatValue, err := ToFloat(DereferenceValue(value))
		if err != nil && IsNilPointerError(err) {
			return floatValue, nil
		}
		return floatValue, err
	}
	valueAsString := AsString(DereferenceValue(value))
	return strconv.ParseFloat(valueAsString, 64)
}

//ToBoolean converts an input to bool.
func ToBoolean(value interface{}) (bool, error) {
	if boolValue, ok := value.(bool); ok {
		return boolValue, nil
	}
	valueAsString := AsString(value)
	return strconv.ParseBool(valueAsString)
}

//AsBoolean converts an input to bool.
func AsBoolean(value interface{}) bool {
	result, err := ToBoolean(value)
	if err != nil {
		return false
	}
	return result
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

var intBitSize = reflect.TypeOf(int64(0)).Bits()

//AsInt converts an input to int.
func AsInt(value interface{}) int {
	var result, err = ToInt(value)
	if err == nil {
		return result
	}
	return 0
}

//ToInt converts input value to int or error
func ToInt(value interface{}) (int, error) {
	if value == nil {
		return 0, NewNilPointerError("int value was nil")
	}
	switch actual := value.(type) {
	case int:
		return actual, nil
	case int8:
		return int(actual), nil
	case int16:
		return int(actual), nil
	case int32:
		return int(actual), nil
	case int64:
		return int(actual), nil
	case uint:
		return int(actual), nil
	case uint8:
		return int(actual), nil
	case uint16:
		return int(actual), nil
	case uint32:
		return int(actual), nil
	case uint64:
		return int(actual), nil
	case float32:
		return int(actual), nil
	case float64:
		return int(actual), nil
	case bool:
		if actual {
			return 1, nil
		}
		return 0, nil
	}
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		value := DereferenceValue(value)
		if intValue, err := ToInt(value); err != nil {
			if err != nil && IsNilPointerError(err) {
				return intValue, err
			}
			return intValue, err
		}
	}
	valueAsString := AsString(value)
	if strings.Contains(valueAsString, ".") {
		floatValue, err := strconv.ParseFloat(valueAsString, intBitSize)
		if err != nil {
			return 0, err
		}
		return int(floatValue), nil
	}
	result, err := strconv.ParseInt(valueAsString, 10, 64)
	return int(result), err
}

func unitToTime(timestamp int64) *time.Time {
	var timeValue time.Time
	if timestamp > math.MaxUint32 {
		var timestampInMs = timestamp / 1000000
		if timestampInMs > math.MaxUint32 {
			timeValue = time.Unix(0, timestamp)
		} else {
			timeValue = time.Unix(0, timestamp*1000000)
		}
	} else {
		timeValue = time.Unix(timestamp, 0)
	}
	return &timeValue
}

func textToTime(value, dateLayout string) (*time.Time, error) {
	floatValue, err := ToFloat(value)
	if err == nil {
		return unitToTime(int64(floatValue)), nil
	}
	timeValue, err := ParseTime(value, dateLayout)
	if err != nil {
		if dateLayout != "" {
			if len(value) > len(dateLayout) {
				value = string(value[:len(dateLayout)])
			}
			timeValue, err = ParseTime(value, dateLayout)
		}
		if err != nil { //JSON default time format fallback
			if timeValue, err = ParseTime(value, time.RFC3339); err == nil {
				return &timeValue, err
			}
			return nil, err
		}
	}
	return &timeValue, nil
}

//ToTime converts value to time, optionally uses layout if value if of string type
func ToTime(value interface{}, dateLayout string) (*time.Time, error) {
	if value == nil {
		return nil, errors.New("values was empty")
	}
	switch actual := value.(type) {
	case time.Time:
		return &actual, nil
	case *time.Time:
		return actual, nil
	case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		var floatValue, _ = ToFloat(value)
		return unitToTime(int64(floatValue)), nil
	case *float32, *float64, *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		actual = DereferenceValue(actual)
		return ToTime(actual, dateLayout)
	case string:
		return textToTime(actual, dateLayout)
	default:
		textValue := AsString(DereferenceValue(actual))
		return textToTime(textValue, dateLayout)
	}
}

//AsTime converts an input to time, it takes time input,  dateLaout as parameters.
func AsTime(value interface{}, dateLayout string) *time.Time {
	result, err := ToTime(value, dateLayout)
	if err != nil {
		return nil
	}
	return result
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
	DateLayout   string
	MappedKeyTag string
}

func (c *Converter) assignConvertedMap(target, source interface{}, targetIndirectValue reflect.Value, targetIndirectPointerType reflect.Type) error {
	mapType := DiscoverTypeByKind(target, reflect.Map)
	mapPointer := reflect.New(mapType)
	mapValueType := mapType.Elem()
	mapKeyType := mapType.Key()
	newMap := mapPointer.Elem()
	newMap.Set(reflect.MakeMap(mapType))
	var err error
	err = ProcessMap(source, func(key, value interface{}) bool {
		if value == nil {
			return true
		}
		mapValueType = reflect.TypeOf(value)
		targetMapValuePointer := reflect.New(mapValueType)
		err = c.AssignConverted(targetMapValuePointer.Interface(), value)
		if err != nil {
			err = fmt.Errorf("failed to assigned converted map value %v to %v due to %v", source, target, err)
			return false
		}

		targetMapKeyPointer := reflect.New(mapKeyType)
		err = c.AssignConverted(targetMapKeyPointer.Interface(), key)
		if err != nil {
			err = fmt.Errorf("failed to assigned converted map key %v to %v due to %v", source, target, err)
			return false
		}
		var elementKey = targetMapKeyPointer.Elem()
		var elementValue = targetMapValuePointer.Elem()

		if elementKey.Type() != mapKeyType {
			if elementKey.Type().AssignableTo(mapKeyType) {
				elementKey = elementKey.Convert(mapKeyType)
			}
		}
		if !elementValue.Type().AssignableTo(newMap.Type().Elem()) {
			var compatibleValue = reflect.New(newMap.Type().Elem())
			err = c.AssignConverted(compatibleValue.Interface(), elementValue.Interface())
			if err != nil {
				return false
			}
			elementValue = compatibleValue.Elem()
		}
		newMap.SetMapIndex(elementKey, elementValue)
		return true
	})
	if err != nil {
		return err
	}
	if targetIndirectPointerType.Kind() == reflect.Map {
		if targetIndirectValue.Type().AssignableTo(mapPointer.Type()) {
			targetIndirectValue.Set(mapPointer)
		} else {
			targetIndirectValue.Set(mapPointer.Elem())
		}
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
			err = fmt.Errorf("failed to convert slice item from %T to %T, values: from %v to %v, due to %v", item, targetComponentPointer.Interface(), item, targetComponentPointer.Interface(), err)
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

	var defaultValueMap = make(map[string]interface{})

	var anonymousValueMap map[string]reflect.Value
	var anonymousFields map[string]reflect.Value

	for _, value := range fieldsMapping {
		if defaultValue, ok := value[defaultKey]; ok {
			var fieldName = value[fieldNameKey]
			defaultValueMap[fieldName] = defaultValue
		}
		if index, ok := value[fieldIndexKey]; ok {
			if len(anonymousValueMap) == 0 {
				anonymousValueMap = make(map[string]reflect.Value)
				anonymousFields = make(map[string]reflect.Value)
			}
			field := newStruct.Field(AsInt(index))
			if field.Type().Kind() == reflect.Ptr {
				fieldStruct := reflect.New(field.Type().Elem())
				anonymousValueMap[index] = fieldStruct
				anonymousFields[index] = field
			} else {
				anonymousValueMap[index] = field.Addr()
				anonymousFields[index] = field.Addr()
			}
		}
	}

	for key, value := range inputMap {
		aStruct := newStruct
		mapping, found := fieldsMapping[strings.ToLower(key)]
		if found {
			var field reflect.Value
			fieldName := mapping[fieldNameKey]
			if fieldIndex, ok := mapping[fieldIndexKey]; ok {
				var structPointer = anonymousValueMap[fieldIndex]
				if anonymousFields[fieldIndex].CanAddr() {
					anonymousFields[fieldIndex].Set(structPointer)
				}
				aStruct = structPointer.Elem()
				initAnonymousStruct(structPointer.Interface())
			}
			field = aStruct.FieldByName(fieldName)
			fieldType, _ := aStruct.Type().FieldByName(fieldName)
			if isExported := fieldType.PkgPath == ""; !isExported {
				structField := &StructField{
					Owner: newStructPointer,
					Value: field,
					Type:  fieldType,
				}
				if !onUnexportedHandler(structField) {
					continue
				}
				field = structField.Value
			}
			if _, has := defaultValueMap[fieldName]; has {
				delete(defaultValueMap, fieldName)
			}
			previousLayout := c.DateLayout
			if HasTimeLayout(mapping) {
				c.DateLayout = GetTimeLayout(mapping)
				c.DateLayout = previousLayout
			}

			if (!field.CanAddr()) && field.Kind() == reflect.Ptr {
				if err := c.AssignConverted(field.Interface(), value); err != nil {
					return fmt.Errorf("failed to convert %v to %v due to %v", value, field, err)
				}

			} else {
				if err := c.AssignConverted(field.Addr().Interface(), value); err != nil {
					return fmt.Errorf("failed to convert %v to %v due to %v", value, field, err)
				}
			}
			if HasTimeLayout(mapping) {
				c.DateLayout = previousLayout
			}
		}
	}

	for fieldName, value := range defaultValueMap {
		field := newStruct.FieldByName(fieldName)
		err := c.AssignConverted(field.Addr().Interface(), value)
		if err != nil {
			return fmt.Errorf("failed to assign default value %v to %v due to %v", value, field, err)
		}
	}

	if targetIndirectPointerType.Kind() == reflect.Slice {
		targetIndirectValue.Set(newStructPointer)
	} else {
		targetIndirectValue.Set(newStruct)
	}
	return nil
}

//customConverter map of target, source type with converter
var customConverter = make(map[reflect.Type]map[reflect.Type]func(target, source interface{}) error)

//RegisterConverter register custom converter for supplied target, source type
func RegisterConverter(target, source reflect.Type, converter func(target, source interface{}) error) {
	if _, ok := customConverter[target]; !ok {
		customConverter[target] = make(map[reflect.Type]func(target, source interface{}) error)
	}
	customConverter[target][source] = converter
}

//GetConverter returns register converter for supplied target and source type
func GetConverter(target, source interface{}) (func(target, source interface{}) error, bool) {
	sourceConverters, ok := customConverter[reflect.TypeOf(target)]
	if !ok {
		return nil, false
	}
	converter, ok := sourceConverters[reflect.TypeOf(source)]
	return converter, ok
}

//AssignConverted assign to the target source, target needs to be pointer, input has to be convertible or compatible type
func (c *Converter) AssignConverted(target, source interface{}) error {
	if target == nil {
		return fmt.Errorf("destination Pointer was nil %v %v", target, source)
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
		case *string:
			transient := []string{*sourceValue}
			*targetValuePointer = transient
			return nil
		case string:
			transient := []string{sourceValue}
			*targetValuePointer = transient
			return nil
		default:
			if IsSlice(source) {
				var stingItems = make([]string, 0)
				ProcessSlice(source, func(item interface{}) bool {
					stingItems = append(stingItems, AsString(item))
					return true
				})
				*targetValuePointer = stingItems
				return nil
			} else if IsMap(source) {
				if len(AsMap(source)) == 0 {
					return nil
				}
			}
			return fmt.Errorf("expected []string but had: %T", source)
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
		var intValue, err = ToInt(source)
		if err != nil {
			return err
		}
		directValue.SetInt(int64(intValue))
		return nil

	case **int, **int8, **int16, **int32, **int64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		var intValue, err = ToInt(source)

		if err != nil {
			if IsNilPointerError(err) {
				return nil
			}
			return err
		}
		switch directType.Kind() {
		case reflect.Int8:
			alignValue := int8(intValue)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&alignValue))
		case reflect.Int16:
			alignValue := int16(intValue)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&alignValue))
		case reflect.Int32:
			alignValue := int32(intValue)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&alignValue))
		case reflect.Int64:
			alignValue := int64(intValue)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&alignValue))
		default:
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&intValue))
		}
		return nil
	case *uint, *uint8, *uint16, *uint32, *uint64:
		directValue := reflect.Indirect(reflect.ValueOf(targetValuePointer))
		value, err := ToInt(source)
		if err != nil {
			return err
		}
		directValue.SetUint(uint64(value))
		return nil
	case **uint, **uint8, **uint16, **uint32, **uint64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		value, err := ToInt(source)
		if !IsNilPointerError(err) && err != nil {
			return err
		}
		switch directType.Kind() {
		case reflect.Uint8:
			alignValue := uint8(value)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&alignValue))
		case reflect.Uint16:
			alignValue := uint16(value)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&alignValue))
		case reflect.Uint32:
			alignValue := uint32(value)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&alignValue))
		case reflect.Uint64:
			alignValue := uint64(value)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&alignValue))
		default:
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&value))
		}
		return nil

	case *float32, *float64:
		directValue := reflect.Indirect(reflect.ValueOf(targetValuePointer))
		value, err := ToFloat(source)
		if err != nil {
			return err
		}
		directValue.SetFloat(value)
		return nil
	case **float32, **float64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()

		value, err := ToFloat(source)
		if err != nil {
			return err
		}
		if directType.Kind() == reflect.Float32 {
			float32Value := float32(value)
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&float32Value))
		} else {
			reflect.ValueOf(targetValuePointer).Elem().Set(reflect.ValueOf(&value))
		}
		return nil
	case *time.Time:
		timeValue, err := ToTime(source, c.DateLayout)
		if err != nil {
			return err
		}
		*targetValuePointer = *timeValue
		return nil
	case **time.Time:
		timeValue, err := ToTime(source, c.DateLayout)
		if err != nil {
			return err
		}
		*targetValuePointer = timeValue
		return nil
	case *interface{}:
		if converter, ok := GetConverter(target, source); ok {
			return converter(target, source)
		}
		(*targetValuePointer) = source
		return nil

	case **interface{}:
		if converter, ok := GetConverter(target, source); ok {
			return converter(target, source)
		}
		(*targetValuePointer) = &source
		return nil

	default:
		if converter, ok := GetConverter(target, source); ok {
			return converter(target, source)
		}
	}

	sourceValue := reflect.ValueOf(source)
	if source == nil || !sourceValue.IsValid() || (sourceValue.CanSet() && sourceValue.IsNil()) {
		return nil
	}
	targetValue := reflect.ValueOf(target)
	targetIndirectValue := reflect.Indirect(targetValue)
	if sourceValue.IsValid() {
		if sourceValue.Type().AssignableTo(targetValue.Type()) {
			targetIndirectValue.Set(sourceValue.Elem())
			return nil
		} else if sourceValue.Type().AssignableTo(targetValue.Type().Elem()) && sourceValue.Kind() == targetValue.Type().Elem().Kind() {
			targetValue.Elem().Set(sourceValue)
			return nil
		}
	}
	var targetIndirectPointerType = reflect.TypeOf(target).Elem()
	if targetIndirectPointerType.Kind() == reflect.Ptr || targetIndirectPointerType.Kind() == reflect.Slice || targetIndirectPointerType.Kind() == reflect.Map {
		targetIndirectPointerType = targetIndirectPointerType.Elem()
	}

	if targetIndirectValue.Kind() == reflect.Slice || targetIndirectPointerType.Kind() == reflect.Slice {

		if sourceValue.Kind() == reflect.Ptr && sourceValue.Elem().Kind() == reflect.Slice {
			sourceValue = sourceValue.Elem()
		}
		if sourceValue.Kind() == reflect.Ptr && sourceValue.IsNil() {
			return nil
		}
		if sourceValue.Kind() == reflect.Slice {
			if targetIndirectValue.Kind() == reflect.Map {
				return c.assignConvertedMap(target, source, targetIndirectValue, targetIndirectPointerType)
			}
			return c.assignConvertedSlice(target, source, targetIndirectValue, targetIndirectPointerType)
		}
	}
	if targetIndirectValue.Kind() == reflect.Map || targetIndirectPointerType.Kind() == reflect.Map {
		sourceKind := DereferenceType(sourceValue.Type()).Kind()
		if sourceKind == reflect.Map {
			return c.assignConvertedMap(target, source, targetIndirectValue, targetIndirectPointerType)
		} else if sourceKind == reflect.Struct {
			if source == nil {
				return nil
			}
			if sourceValue.Kind() == reflect.Ptr && sourceValue.IsNil() {
				return nil
			}
			targetValue := reflect.ValueOf(target)
			if !targetValue.CanInterface() {
				return nil
			}
			return c.assignConvertedMapFromStruct(source, target, sourceValue)

		} else if sourceKind == reflect.Slice {
			componentType := DereferenceType(DiscoverComponentType(source))
			if componentType.ConvertibleTo(reflect.TypeOf(keyValue{})) {
				return c.assignConvertedStructSliceToMap(target, source)
			} else if componentType.Kind() == reflect.Map {
				return c.assignConvertedMapSliceToMap(target, source)
			} else if componentType.Kind() == reflect.Interface {
				return c.assignConvertedMapSliceToMap(target, source)
			}
		}

	} else if targetIndirectValue.Kind() == reflect.Struct {
		timeValuePtr := tryExtractTime(target)
		if timeValuePtr != nil {
			if timeValue, err := ToTime(source, c.DateLayout); err == nil {
				return c.assignEmbeddedTime(target, timeValue)
			}
		}

		sourceMap, err := ToMap(source)
		if err != nil {
			return fmt.Errorf("unable to convert %T to %T", source, target)
		}
		return c.assignConvertedStruct(target, sourceMap, targetIndirectValue, targetIndirectPointerType)

	} else if targetIndirectPointerType.Kind() == reflect.Struct {
		structPointer := reflect.New(targetIndirectPointerType)
		inputMap, err := ToMap(source)
		if err != nil {
			return fmt.Errorf("unable transfer to %T,  source should be a map but was %T(%v)", target, source, source)
		}
		if err = c.assignConvertedStruct(target, inputMap, structPointer.Elem(), targetIndirectPointerType); err != nil {
			return err
		}
		targetIndirectValue.Set(structPointer)
		return nil

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

	targetDeRefType := DereferenceType(target)

	for _, candidate := range numericTypes {
		if candidate.Kind() == targetDeRefType.Kind() {
			var pointerCount = CountPointers(target)
			var compatibleTarget = reflect.New(candidate)
			for i := 0; i < pointerCount-1; i++ {
				compatibleTarget = reflect.New(compatibleTarget.Type())
			}
			if err := c.AssignConverted(compatibleTarget.Interface(), source); err == nil {
				targetValue := reflect.ValueOf(target)
				targetValue.Elem().Set(compatibleTarget.Elem().Convert(targetValue.Elem().Type()))
				return nil
			}

		}
	}
	return fmt.Errorf("Unable to convert type %T into type %T\n\t%v", source, target, source)
}

func (c *Converter) assignEmbeddedTime(target interface{}, source *time.Time) error {
	targetValue := reflect.ValueOf(target)
	structValue := targetValue
	if targetValue.Kind() == reflect.Ptr {
		structValue = reflect.Indirect(targetValue)
	}
	anonymous := structValue.Field(0)
	anonymous.Set(reflect.ValueOf(*source))
	if targetValue.Kind() == reflect.Ptr {
		targetValue.Elem().Set(structValue)
	}
	return nil
}

type keyValue struct {
	Key, Value interface{}
}

func (c *Converter) assignConvertedStructSliceToMap(target, source interface{}) (err error) {
	mapType := DiscoverTypeByKind(target, reflect.Map)
	mapPointer := reflect.ValueOf(target)
	mapValueType := mapType.Elem()
	mapKeyType := mapType.Key()
	newMap := mapPointer.Elem()
	newMap.Set(reflect.MakeMap(mapType))
	keyValueType := reflect.TypeOf(keyValue{})
	ProcessSlice(source, func(item interface{}) bool {
		if item == nil {
			return true
		}
		item = reflect.ValueOf(DereferenceValue(item)).Convert(keyValueType).Interface()
		pair, ok := item.(keyValue)
		if !ok {
			return true
		}

		targetMapValuePointer := reflect.New(mapValueType)
		err = c.AssignConverted(targetMapValuePointer.Interface(), pair.Value)
		if err != nil {
			return false
		}
		targetMapKeyPointer := reflect.New(mapKeyType)
		err = c.AssignConverted(targetMapKeyPointer.Interface(), pair.Key)
		if err != nil {
			return false
		}
		var elementKey = targetMapKeyPointer.Elem()
		var elementValue = targetMapValuePointer.Elem()

		if elementKey.Type() != mapKeyType {
			if elementKey.Type().AssignableTo(mapKeyType) {
				elementKey = elementKey.Convert(mapKeyType)
			}
		}
		if !elementValue.Type().AssignableTo(newMap.Type().Elem()) {
			var compatibleValue = reflect.New(newMap.Type().Elem())
			err = c.AssignConverted(compatibleValue.Interface(), elementValue.Interface())
			if err != nil {
				return false
			}
			elementValue = compatibleValue.Elem()
		}
		newMap.SetMapIndex(elementKey, elementValue)
		return true
	})
	return err
}

//entryMapToKeyValue converts entry map into map
func entryMapToKeyValue(entryMap map[string]interface{}) (key string, value interface{}, err error) {
	if len(entryMap) > 2 {
		return key, value, fmt.Errorf("map entry needs to have 2 elements but had: %v, %v", len(entryMap), entryMap)
	}

	hasValue := false
	for k, v := range entryMap {
		if strings.ToLower(k) == "key" {
			key = AsString(v)
			continue
		} else if strings.ToLower(k) == "value" {
			hasValue = true
			value = v
		}
	}
	if key == "" {
		return key, value, fmt.Errorf("key is required in entryMap %v", entryMap)
	}
	if !hasValue && len(entryMap) == 2 {
		return key, value, fmt.Errorf("map entry needs to have key, value pair but had:  %v", entryMap)
	}
	return key, value, nil
}

func (c *Converter) assignConvertedMapSliceToMap(target, source interface{}) (err error) {
	mapType := DiscoverTypeByKind(target, reflect.Map)
	mapPointer := reflect.ValueOf(target)
	mapValueType := mapType.Elem()
	mapKeyType := mapType.Key()
	newMap := mapPointer.Elem()
	newMap.Set(reflect.MakeMap(mapType))
	ProcessSlice(source, func(item interface{}) bool {
		if item == nil {
			return true
		}
		entryMap := AsMap(item)
		key, value, e := entryMapToKeyValue(entryMap)
		if e != nil {
			err = fmt.Errorf("unable to cast %T to %T", source, target)
			return false
		}
		targetMapValuePointer := reflect.New(mapValueType)
		err = c.AssignConverted(targetMapValuePointer.Interface(), value)
		if err != nil {
			return false
		}
		targetMapKeyPointer := reflect.New(mapKeyType)
		err = c.AssignConverted(targetMapKeyPointer.Interface(), key)
		if err != nil {
			return false
		}
		var elementKey = targetMapKeyPointer.Elem()
		var elementValue = targetMapValuePointer.Elem()

		if elementKey.Type() != mapKeyType {
			if elementKey.Type().AssignableTo(mapKeyType) {
				elementKey = elementKey.Convert(mapKeyType)
			}
		}
		if !elementValue.Type().AssignableTo(newMap.Type().Elem()) {
			var compatibleValue = reflect.New(newMap.Type().Elem())
			err = c.AssignConverted(compatibleValue.Interface(), elementValue.Interface())
			if err != nil {
				return false
			}
			elementValue = compatibleValue.Elem()
		}
		newMap.SetMapIndex(elementKey, elementValue)
		return true
	})
	return err
}

func (c *Converter) assignConvertedMapFromStruct(source, target interface{}, sourceValue reflect.Value) error {
	if source == nil || !sourceValue.IsValid() {
		return nil
	}
	targetMap := AsMap(target)
	if targetMap == nil {
		return fmt.Errorf("target %T is not a map", target)
	}

	return ProcessStruct(source, func(fieldType reflect.StructField, field reflect.Value) error {
		if !field.CanInterface() {
			return nil
		}
		value := field.Interface()
		if value == nil {
			return nil
		}
		if timeVal := tryExtractTime(value); timeVal != nil {
			value = timeVal.Format(time.RFC3339)
		}
		var fieldTarget interface{}
		if IsStruct(value) {
			aMap := make(map[string]interface{})
			if err := c.AssignConverted(&aMap, value); err != nil {
				return err
			}
			fieldTarget = aMap

		} else if IsSlice(value) {
			var componentType = DereferenceType(DiscoverComponentType(value))
			if componentType.Kind() == reflect.Struct {
				var slice = make([]map[string]interface{}, 0)
				if err := c.AssignConverted(&slice, value); err != nil {
					return err
				}
				fieldTarget = slice
			} else {
				if _, isByteArray := value.([]byte); isByteArray {
					fieldTarget = value
				} else {
					var slice = make([]interface{}, 0)
					if err := c.AssignConverted(&slice, value); err != nil {
						return err
					}
					fieldTarget = slice
				}
			}
		} else if err := c.AssignConverted(&fieldTarget, value); err != nil {
			return err
		}

		fieldName := fieldType.Name
		keyTag := strings.Trim(fieldType.Tag.Get(c.MappedKeyTag), `"`)

		if keyTag != "" {
			key := strings.Split(keyTag, ",")[0]
			if key == "-" {
				return nil
			}
			fieldName = key
		}
		targetMap[fieldName] = fieldTarget
		return nil
	})
}

func tryExtractTime(value interface{}) *time.Time {

	if timeVal, ok := value.(time.Time); ok {
		return &timeVal
	}
	if timeVal, ok := value.(*time.Time); ok && timeVal != nil {
		return timeVal
	}
	if !IsStruct(value) {
		return nil
	}

	structOrPtrValue := reflect.ValueOf(value)
	structValue := structOrPtrValue
	if structOrPtrValue.Kind() == reflect.Ptr {
		structValue = reflect.Indirect(structOrPtrValue)
	}
	if structValue.Kind() == reflect.Ptr {
		structValue = reflect.ValueOf(DereferenceValue(structValue.Interface()))
	}

	if !structValue.IsValid() {
		return nil
	}
	if structValue.NumField() > 1 {
		return nil
	}
	timeField, ok := structValue.Type().FieldByName("Time")
	if !ok || !timeField.Anonymous {
		return nil
	}
	timeValue := structValue.Field(timeField.Index[0])
	if timeValue.CanAddr() {
		return tryExtractTime(timeValue.Addr().Interface())
	}
	return tryExtractTime(timeValue.Interface())
}

//NewColumnConverter create a new converter, that has ability to convert map to struct using column mapping
func NewColumnConverter(dateLayout string) *Converter {
	return &Converter{dateLayout, "column"}
}

//NewConverter create a new converter, that has ability to convert map to struct, it uses keytag to identify source and dest of fields/keys
func NewConverter(dateLayout, keyTag string) *Converter {
	if keyTag == "" {
		keyTag = "name"
	}
	return &Converter{dateLayout, keyTag}
}

//DefaultConverter represents a default data structure converter
var DefaultConverter = NewConverter("", "name")

//DereferenceValues replaces pointer to its value within a generic  map or slice
func DereferenceValues(source interface{}) interface{} {
	if IsMap(source) {
		var aMap = make(map[string]interface{})
		_ = ProcessMap(source, func(key, value interface{}) bool {
			if value == nil {
				return true
			}
			aMap[AsString(key)] = DereferenceValue(value)

			return true
		})
		return aMap

	} else if IsSlice(source) {
		var aSlice = make([]interface{}, 0)
		ProcessSlice(source, func(item interface{}) bool {
			aSlice = append(aSlice, DereferenceValue(item))
			return true
		})
		return aSlice

	}
	return DereferenceValue(source)
}

//DereferenceValue dereference passed in value
func DereferenceValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	var reflectValue reflect.Value
	switch actualValue := value.(type) {
	case reflect.Value:
		reflectValue = actualValue
	default:
		reflectValue = reflect.ValueOf(value)
	}
	for {
		if !reflectValue.IsValid() {
			break
		}

		if !reflectValue.CanInterface() {
			break
		}
		if reflectValue.Type().Kind() != reflect.Ptr {
			break
		}

		reflectValue = reflectValue.Elem()
	}

	var result interface{}
	if reflectValue.IsValid() && reflectValue.CanInterface() {
		result = reflectValue.Interface()
	}

	if result != nil && (IsMap(result) || IsSlice(result)) {
		return DereferenceValues(value)
	}
	return result
}

//DereferenceType dereference passed in value
func DereferenceType(value interface{}) reflect.Type {
	if value == nil {
		return nil
	}
	var reflectType reflect.Type
	reflectValue, ok := value.(reflect.Value)
	if ok {
		reflectType = reflectValue.Type()
	} else if reflectType, ok = value.(reflect.Type); !ok {
		reflectType = reflect.TypeOf(value)
	}

	for {
		if reflectType.Kind() != reflect.Ptr {
			break
		}
		reflectType = reflectType.Elem()
	}

	return reflectType
}

//CountPointers count pointers to undelying non pointer type
func CountPointers(value interface{}) int {
	if value == nil {
		return 0
	}
	var result = 0
	reflectType, ok := value.(reflect.Type)
	if !ok {
		reflectType = reflect.TypeOf(value)
	}

	for {
		if reflectType.Kind() != reflect.Ptr {
			break
		}
		result++
		reflectType = reflectType.Elem()
	}

	return result
}

func initAnonymousStruct(aStruct interface{}) {
	structValue := DiscoverValueByKind(reflect.ValueOf(aStruct), reflect.Struct)
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		if !fieldType.Anonymous {
			continue
		}
		field := structValue.Field(i)
		if !IsStruct(field) {
			continue
		}

		var aStruct interface{}
		if fieldType.Type.Kind() == reflect.Ptr {
			if field.IsNil() {
				if !field.CanSet() {
					continue
				}
				structValue.Field(i).Set(reflect.New(fieldType.Type.Elem()))
			}
			aStruct = field.Interface()
		} else {
			if !field.CanAddr() {
				continue
			}
			aStruct = field.Addr().Interface()
		}
		initAnonymousStruct(aStruct)
	}
}
