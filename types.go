package toolbox

import (
	"fmt"
	"reflect"
	"time"
)

//Zeroable represents object that can call IsZero
type Zeroable interface {
	//IsZero returns true, if value of object was zeroed.
	IsZero() bool
}

//IsInt returns true if input is an int
func IsInt(input interface{}) bool {
	switch input.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

//IsNumber returns true if type is either float or int
func IsNumber(input interface{}) bool {
	return IsFloat(input) || IsInt(input)
}

//IsFloat returns true if input is a float
func IsFloat(input interface{}) bool {
	switch input.(type) {
	case float32, float64:
		return true
	}
	return false
}

//IsBool returns true if input is a boolean
func IsBool(input interface{}) bool {
	switch input.(type) {
	case bool:
		return true
	}
	return false
}

//IsString returns true if input is a string
func IsString(input interface{}) bool {
	switch input.(type) {
	case string:
		return true
	}
	return false
}

//CanConvertToString checks if  input can be converted to string
func CanConvertToString(input interface{}) bool {
	return reflect.TypeOf(input).AssignableTo(reflect.TypeOf(""))
}

//IsTime returns true if input is a time
func IsTime(input interface{}) bool {
	switch input.(type) {
	case time.Time:
		return true
	case *time.Time:
		return true
	}
	return false
}

//IsMap returns true if input is a map
func IsMap(input interface{}) bool {
	switch input.(type) {
	case map[string]interface{}:
		return true
	}
	candidateType := DereferenceType(reflect.TypeOf(input))
	return candidateType.Kind() == reflect.Map
}

//IsStruct returns true if input is a map
func IsStruct(input interface{}) bool {
	if input == nil {
		return false
	}
	inputType := DereferenceType(input)
	return inputType.Kind() == reflect.Struct
}

//IsSlice returns true if input is a map
func IsSlice(input interface{}) bool {
	switch input.(type) {
	case []interface{}:
		return true
	case []string:
		return true
	}
	candidateType := DereferenceType(reflect.TypeOf(input))
	return candidateType.Kind() == reflect.Slice
}

//IsFunc returns true if input is a funct
func IsFunc(input interface{}) bool {
	candidateType := DereferenceType(reflect.TypeOf(input))
	return candidateType.Kind() == reflect.Func
}

//IsZero returns true if input is a zeroable
func IsZero(input interface{}) bool {
	if zeroable, ok := input.(Zeroable); ok {
		return zeroable.IsZero()
	}
	return false
}

//IsPointer returns true if input is a pointer
func IsPointer(input interface{}) bool {
	if reflectType, ok := input.(reflect.Type); ok {
		return reflectType.Kind() == reflect.Ptr
	}
	return reflect.TypeOf(input).Kind() == reflect.Ptr
}

//AssertPointerKind checks if input is a pointer of the passed in kind, if not it panic with message including name
func AssertPointerKind(input interface{}, kind reflect.Kind, name string) {
	AssertTypeKind(reflect.TypeOf(input), reflect.Ptr, name)
	AssertTypeKind(reflect.TypeOf(input).Elem(), kind, name)
}

//AssertKind checks if input is of the passed in kind, if not it panic with message including name
func AssertKind(input interface{}, kind reflect.Kind, name string) {
	AssertTypeKind(reflect.TypeOf(input), kind, name)
}

//AssertTypeKind checks if dataType is of the passed in kind, if not it panic with message including name
func AssertTypeKind(dataType reflect.Type, kind reflect.Kind, name string) {
	if dataType.Kind() != kind {
		panic(fmt.Sprintf("failed to check: %v - expected kind: %v but found %v (%v)", name, kind.String(), dataType.Kind(), dataType.String()))
	}
}

//DiscoverValueByKind returns unwrapped input that matches expected kind, or panic if this is not possible
func DiscoverValueByKind(input interface{}, expected reflect.Kind) reflect.Value {
	result, err := TryDiscoverValueByKind(input, expected)
	if err == nil {
		return result
	}
	panic(err)
}

//TryDiscoverValueByKind returns unwrapped input that matches expected kind, or panic if this is not possible
func TryDiscoverValueByKind(input interface{}, expected reflect.Kind) (reflect.Value, error) {
	value, ok := input.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(input)
	}
	if value.Kind() == expected {
		return value, nil
	} else if value.Kind() == reflect.Ptr {
		return TryDiscoverValueByKind(value.Elem(), expected)
	} else if value.Kind() == reflect.Interface {
		return TryDiscoverValueByKind(value.Elem(), expected)
	}
	return value, fmt.Errorf("failed to discover value by kind expected: %v, actual:%T   on %v:", expected.String(), value.Type(), value)
}

//IsValueOfKind returns true if passed in input is of supplied kind.
func IsValueOfKind(input interface{}, kind reflect.Kind) bool {
	value, ok := input.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(input)
	}
	if value.Kind() == kind {
		return true
	} else if value.Kind() == reflect.Ptr {
		return IsValueOfKind(value.Elem(), kind)
	} else if value.Kind() == reflect.Interface {
		return IsValueOfKind(value.Elem(), kind)
	}
	return false
}

//DiscoverTypeByKind returns unwrapped input type that matches expected kind, or panic if this is not possible
func DiscoverTypeByKind(input interface{}, expected reflect.Kind) reflect.Type {
	result, err := TryDiscoverTypeByKind(input, expected)
	if err != nil {
		panic(err)
	}
	return result
}

//TryDiscoverTypeByKind returns unwrapped input type that matches expected kind, or error
func TryDiscoverTypeByKind(input interface{}, expected reflect.Kind) (reflect.Type, error) {
	value, ok := input.(reflect.Type)
	if !ok {
		value = reflect.TypeOf(input)
	}
	if value.Kind() == expected {
		return value, nil
	} else if value.Kind() == reflect.Ptr || value.Kind() == reflect.Slice {
		return TryDiscoverTypeByKind(value.Elem(), expected)
	}
	return nil, fmt.Errorf("failed to discover type by kind %v, on %v:", expected.String(), value)
}

//DiscoverComponentType returns type unwrapped from pointer, slice or map
func DiscoverComponentType(input interface{}) reflect.Type {
	valueType, ok := input.(reflect.Type)
	if !ok {
		valueType = reflect.TypeOf(input)
	}
	if valueType.Kind() == reflect.Ptr {
		return DiscoverComponentType(valueType.Elem())
	} else if valueType.Kind() == reflect.Slice {
		return valueType.Elem()
	} else if valueType.Kind() == reflect.Map {
		return valueType.Elem()
	}
	return valueType
}
