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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/viant/toolbox"
	"github.com/stretchr/testify/assert"
)

type User2 struct {
	Name        string    `column:"name"`
	DateOfBirth time.Time `column:"date" dateLayout:"2006-01-02 15:04:05.000000"`
	Id          int       `autoincrement:"true"`
	Other       string    `transient:"true"`
}


func TestAssertKind(t *testing.T) {
	toolbox.AssertKind(User2{}, reflect.Struct, "user")
	toolbox.AssertKind((*User2)(nil), reflect.Ptr, "user")

	defer func() {
		if err := recover(); err != nil {
			expected := "Failed to check: User - expected kind: ptr but found struct (toolbox_test.User2)"
			actual := fmt.Sprintf("%v", err)
			assert.Equal(t, actual, expected, "Assert Kind")
		}
	}()
	toolbox.AssertKind(User2{}, reflect.Ptr, "User")
}

func TestAssertPointerKind(test *testing.T) {
	toolbox.AssertPointerKind(&User2{}, reflect.Struct, "user")
	toolbox.AssertPointerKind((*User2)(nil), reflect.Struct, "user")
}





func TestTypeDetection(t *testing.T) {

	assert.False(t, toolbox.IsFloat(3))
	assert.True(t, toolbox.IsFloat(3.0))
	assert.True(t, toolbox.CanConvertToFloat(3.0))
	assert.True(t, toolbox.CanConvertToFloat("3"))
	assert.True(t, toolbox.CanConvertToFloat(3))
	assert.False(t, toolbox.CanConvertToFloat(false))


	assert.False(t, toolbox.IsInt(3.0))
	assert.True(t, toolbox.IsInt(3))

	assert.True(t, toolbox.CanConvertToInt(3))
	assert.True(t, toolbox.CanConvertToInt("3"))


	assert.False(t, toolbox.CanConvertToInt(true))
	assert.False(t, toolbox.CanConvertToInt(3.3))


	assert.False(t, toolbox.IsBool(3.0))
	assert.True(t, toolbox.IsBool(true))

	assert.False(t, toolbox.IsString(3.0))
	assert.True(t, toolbox.IsString("abc"))


	assert.True(t, toolbox.CanConvertToString("abc"))
	assert.False(t, toolbox.CanConvertToString(3.2))

	assert.False(t, toolbox.IsTime(3.0))
	assert.True(t, toolbox.IsTime(time.Now()))
	var timeValues = make([]time.Time, 1)
	assert.True(t, toolbox.IsZero(timeValues[0]))
	assert.False(t, toolbox.IsZero(time.Now()))

	assert.False(t, toolbox.IsZero(""))

	aString := ""
	assert.True(t, toolbox.IsPointer(&aString))
	assert.False(t, toolbox.IsPointer(aString))



}


/*





//IsPointer returns true if input is a pointer

//AssertPointerKind checks if input is a pointer of the passed in kind, if not it panic with message including name
func AssertPointerKind(input interface{}, kind reflect.Kind, name string) {
	AssertType(reflect.TypeOf(input), reflect.Ptr, name)
	AssertType(reflect.TypeOf(input).Elem(), kind, name)
}

//AssertKind checks if input is of the passed in kind, if not it panic with message including name
func AssertKind(input interface{}, kind reflect.Kind, name string) {
	AssertType(reflect.TypeOf(input), kind, name)
}

//AssertType checks if dataType is of the passed in kind, if not it panic with message including name
func AssertType(dataType reflect.Type, kind reflect.Kind, name string) {
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

 */