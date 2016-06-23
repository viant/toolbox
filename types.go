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

//IsString returns true if input is an string
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
	}
	return false
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

//AssertType checks if dataType is of the passed in kind, if not it panic with message including name
func AssertTypeKind(dataType reflect.Type, kind reflect.Kind, name string) {
	if dataType.Kind() != kind {
		panic(fmt.Sprintf("Failed to check: %v - expected kind: %v but found %v (%v)", name, kind.String(), dataType.Kind(), dataType.String()))
	}
}

//DiscoverValueByKind returns unwrapped input that matches expected kind, or panic if this is not possible
func DiscoverValueByKind(input interface{}, expected reflect.Kind) reflect.Value {
	value, ok := input.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(input)
	}
	if value.Kind() == expected {
		return value
	} else if value.Kind() == reflect.Ptr {
		return DiscoverValueByKind(value.Elem(), expected)
	} else if value.Kind() == reflect.Interface {
		return DiscoverValueByKind(value.Elem(), expected)
	}
	panic(fmt.Sprintf("Failed to discover value by kind expected: %v, actual:%v   on %v:", expected.String(), value.Type(), value))
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
	value, ok := input.(reflect.Type)
	if !ok {
		value = reflect.TypeOf(input)
	}
	if value.Kind() == expected {
		return value
	} else if value.Kind() == reflect.Ptr || value.Kind() == reflect.Slice {
		return DiscoverTypeByKind(value.Elem(), expected)
	}
	panic(fmt.Sprintf("Failed to discover type by kind %v, on %v:", expected.String(), value))
}

//DiscoverComponentType returns type unwrapped from pointer, slice or map
func DiscoverComponentType(input interface{}) reflect.Type {
	value, ok := input.(reflect.Type)
	if !ok {
		value = reflect.TypeOf(input)
	}
	if value.Kind() == reflect.Ptr {
		return DiscoverComponentType(value.Elem())
	} else if value.Kind() == reflect.Slice {
		return value.Elem()
	} else if value.Kind() == reflect.Map {
		return value.Elem()
	}
	return value
}
