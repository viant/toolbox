package toolbox

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"errors"
	"math"
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
	if result, err := ToFloat(value); err == nil {
		return result
	}
	return 0
}

//ToFloat converts an input to float or error
func ToFloat(value interface{}) (float64, error) {
	switch actualValue := value.(type) {
	case float64:
		return actualValue, nil
	case float32:
		return float64(actualValue), nil
	case *float64:
		return *actualValue, nil
	}
	valueAsString := AsString(value)
	return strconv.ParseFloat(valueAsString, 64)
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
	case *int:
		return *actual, nil
	case *int8:
		return int(*actual), nil
	case *int16:
		return int(*actual), nil
	case *int32:
		return int(*actual), nil
	case *int64:
		return int(*actual), nil
	case *uint:
		return int(*actual), nil
	case *uint8:
		return int(*actual), nil
	case *uint16:
		return int(*actual), nil
	case *uint32:
		return int(*actual), nil
	case *uint64:
		return int(*actual), nil
	case *float32:
		return int(*actual), nil
	case *float64:
		return int(*actual), nil
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



func ToTime(value interface{}, dateLayout string) (*time.Time, error) {
	if value == nil{
		return nil, errors.New("values was empty")
	}
	switch actual := value.(type) {
	case time.Time:
		return &actual, nil
	case *time.Time:
		return actual, nil
	default:
		floatValue, err := ToFloat(value)
		if err == nil {
			var timeValue time.Time
			var timestamp = int64(floatValue)
			if timestamp > math.MaxUint32 {
				var timestampInMs =  timestamp / 1000000
				if timestampInMs > math.MaxUint32 {
					timeValue = time.Unix(0, timestamp)
				} else {
					timeValue = time.Unix(0, timestamp* 1000000)
				}
			}  else {
				timeValue = time.Unix(timestamp, 0)
			}
			return &timeValue, nil
		}
		timeValue, err := ParseTime(AsString(value), dateLayout)
		if err != nil {
			return nil, err
		}
		return &timeValue, nil
	}
	return nil, fmt.Errorf("unsupported type: %T", value)
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
	DataLayout   string
	MappedKeyTag string
}

func (c *Converter) assignConvertedMap(target, input interface{}, targetIndirectValue reflect.Value, targetIndirectPointerType reflect.Type) error {
	mapType := DiscoverTypeByKind(target, reflect.Map)
	mapPointer := reflect.New(mapType)
	mapValueType := mapType.Elem()
	mapKeyType := mapType.Key()
	newMap := mapPointer.Elem()
	newMap.Set(reflect.MakeMap(mapType))
	var err error

	ProcessMap(input, func(key, value interface{}) bool {
		if value == nil {
			return true
		}
		mapValueType = reflect.TypeOf(value)
		targetMapValuePointer := reflect.New(mapValueType)
		err = c.AssignConverted(targetMapValuePointer.Interface(), value)
		if err != nil {
			err = fmt.Errorf("failed to assigned converted map value %v to %v due to %v", input, target, err)
			return false
		}

		targetMapKeyPointer := reflect.New(mapKeyType)
		err = c.AssignConverted(targetMapKeyPointer.Interface(), key)
		if err != nil {
			err = fmt.Errorf("failed to assigned converted map key %v to %v due to %v", input, target, err)
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

	if targetIndirectPointerType.Kind() == reflect.Map {
		targetIndirectValue.Set(mapPointer.Elem())
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
					return fmt.Errorf("failed to convert %v to %v due to %v", value, field, err)
				}
				c.DataLayout = previousLayout

			} else {

				err := c.AssignConverted(field.Addr().Interface(), value)
				if err != nil {
					return fmt.Errorf("failed to convert %v to %v due to %v", value, field, err)
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
		var intValue, err = ToInt(DereferenceValue(source))
		if err != nil {
			return err
		}
		directValue.SetInt(int64(intValue))
		return nil

	case **int, **int8, **int16, **int32, **int64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		var intValue, err = ToInt(DereferenceValue(source))
		if err != nil {
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
		value, err := ToInt(DereferenceValue(source))
		if err != nil {
			return err
		}
		directValue.SetUint(uint64(value))
		return nil
	case **uint, **uint8, **uint16, **uint32, **uint64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()
		value, err := ToInt(DereferenceValue(source))
		if err != nil {
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
		value, err := ToFloat(DereferenceValue(source))
		if err != nil {
			return err
		}
		directValue.SetFloat(value)
		return nil
	case **float32, **float64:
		directType := reflect.TypeOf(targetValuePointer).Elem().Elem()

		value, err := ToFloat(DereferenceValue(source))
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
		sourceKind := DereferenceType(sourceValue.Type()).Kind()
		if sourceKind == reflect.Map {
			return c.assignConvertedMap(target, source, targetIndirectValue, targetIndirectPointerType)
		} else if sourceKind == reflect.Struct {
			sourceValue = reflect.ValueOf(DereferenceValue(source))
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
		if !IsMap(source) {
			return fmt.Errorf("unable transfer to %T,  source should be a map but was %T(%v)", target, source, source)
		}
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

	targetDereferecedType := DereferenceType(target)

	for _, candidate := range numericTypes {
		if candidate.Kind() == targetDereferecedType.Kind() {
			var pointerCount = CountPointers(target)
			var compatibleTarget = reflect.New(candidate)
			for i := 0; i < pointerCount-1; i++ {
				compatibleTarget = reflect.New(compatibleTarget.Type())
			}
			c.AssignConverted(compatibleTarget.Interface(), source)
			targetValue := reflect.ValueOf(target)
			targetValue.Elem().Set(compatibleTarget.Elem().Convert(targetValue.Elem().Type()))
			return nil
		}
	}
	return fmt.Errorf("Unable to convert type %T into type %T\n\t%v", source, target, source)
}

func (c *Converter) assignConvertedMapFromStruct(source, target interface{}, sourceValue reflect.Value) error {
	targetMap := AsMap(target)
	if targetMap == nil {
		return fmt.Errorf("target %T is not a map", target)
	}
	if source == nil || !sourceValue.IsValid() {
		return nil
	}

	sourceType := sourceValue.Type()

	for i := 0; i < sourceValue.NumField(); i++ {
		field := sourceValue.Field(i)

		if ! field.CanInterface() {
			continue
		}
		var value interface{}
		fieldKind := DereferenceType(field.Type()).Kind()
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

//DereferenceValues replaces pointer to its value within a generic  map or slice
func DereferenceValues(source interface{}) interface{} {
	if IsMap(source) {
		var aMap = make(map[string]interface{})
		ProcessMap(source, func(key, value interface{}) bool {
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
	reflectType, ok := value.(reflect.Type)
	if !ok {
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
