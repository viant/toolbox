package toolbox_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
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
			expected := "failed to check: User - expected kind: ptr but found struct (toolbox_test.User2)"
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
	assert.False(t, toolbox.IsPointer(reflect.TypeOf(aString)))
	assert.True(t, toolbox.IsPointer(&aString))

}

func TestIsValueOfKind(t *testing.T) {
	text := ""
	assert.True(t, toolbox.IsValueOfKind(&text, reflect.Ptr))
	assert.False(t, toolbox.IsValueOfKind(&text, reflect.Struct))
	assert.True(t, toolbox.IsValueOfKind(&text, reflect.String))

	values := make([]interface{}, 1)
	values[0] = 1
	assert.True(t, toolbox.IsValueOfKind(&values[0], reflect.Int))

}

func TestIsFunc(t *testing.T) {
	var f = func() {}
	assert.True(t, toolbox.IsFunc(&f))
	assert.True(t, toolbox.IsFunc(f))
	assert.False(t, toolbox.IsFunc(""))

}
