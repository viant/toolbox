package unsafe

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestFieldPointer(t *testing.T) {
	var testCases = []struct {
		description string
		instance interface{}
		fieldIndex int
		expect interface{}
	}{
		{
			description: "int type",
			instance: struct {
				A []int
				B int
			}{
				nil, 101,
			},
			fieldIndex: 1,
			expect:     101,
		},
		{
			description: "[]byte type",
			instance: struct {
				A []int
				B int
				C []byte
			}{
				nil, 102, []byte{'a', 'c', 'b'},
			},
			fieldIndex: 2,
			expect:     []byte{'a', 'c', 'b'},
		},
		{
			description: "*bool type",
			instance: struct {
				A []int
				B int
				C []byte
				D *bool
			}{
				nil, 102, []byte{'a', 'c', 'b'}, pBool(true),
			},
			fieldIndex: 3,
			expect:     true,
		},
		{
			description: "*[]string type",
			instance: struct {
				A []int
				B int
				C []byte
				D *bool
				E *[]string
			}{
				nil, 102, []byte{'a', 'c', 'b'}, pBool(true),pStrings([]string{"a", "1", "a"}),
			},
			fieldIndex: 4,
			expect:     []string{"a", "1", "a"},
		},
	}


	for _, testCase := range testCases {
		structValue :=  reflect.ValueOf(testCase.instance)
		structPtr := reflect.New(structValue.Type())
		structPtr.Elem().Set(structValue)

		ptr, err := FieldPointer(structValue.Type(), testCase.fieldIndex)
		if ! assert.Nil(t, err, testCase.description) {
			continue
		}
		holderAddr := structPtr.Elem().UnsafeAddr()
		actual := dereference(reflect.ValueOf(ptr(holderAddr))).Interface()
		assert.EqualValues(t, testCase.expect, actual, testCase.description)
	}

}

func dereference(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return dereference(v.Elem())
	}
	return v
}

func pBool(b bool) *bool {
	return &b
}

func pStrings(s []string) *[]string {
	return &s
}