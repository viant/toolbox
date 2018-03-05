package toolbox

import (
	"reflect"
	"time"
)

//Iterator represents generic iterator.
type Iterator interface {
	//HasNext returns true if iterator has next element.
	HasNext() bool

	//Next sets item pointer with next element.
	Next(itemPointer interface{}) error
}

type sliceIterator struct {
	sliceValue reflect.Value
	index      int
}

func (i *sliceIterator) HasNext() bool {
	return i.index < i.sliceValue.Len()
}

func (i *sliceIterator) Next(itemPointer interface{}) error {
	value := i.sliceValue.Index(i.index)
	i.index++
	itemPointerValue := reflect.ValueOf(itemPointer)
	itemPointerValue.Elem().Set(value)
	return nil
}

type stringSliceIterator struct {
	sliceValue []string
	index      int
}

func (i *stringSliceIterator) HasNext() bool {
	return i.index < len(i.sliceValue)
}

func (i *stringSliceIterator) Next(itemPointer interface{}) error {
	value := i.sliceValue[i.index]
	i.index++
	if stringPointer, ok := itemPointer.(*string); ok {
		*stringPointer = value
		return nil
	}
	interfacePointer := itemPointer.(*interface{})
	*interfacePointer = value
	return nil
}

type interfaceSliceIterator struct {
	sliceValue []interface{}
	index      int
}

func (i *interfaceSliceIterator) HasNext() bool {
	return i.index < len(i.sliceValue)
}

func (i *interfaceSliceIterator) Next(itemPointer interface{}) error {
	value := i.sliceValue[i.index]
	i.index++

	switch actual := itemPointer.(type) {
	case *interface{}:
		*actual = value
		return nil
	case *string:
		*actual = AsString(value)
		return nil
	case *int:
		*actual = AsInt(value)
		return nil
	case *int64:
		*actual = int64(AsInt(value))
		return nil
	case *time.Time:
		var timestamp = AsTime(value, DefaultDateLayout)
		if timestamp != nil {
			*actual = *timestamp
		}
		return nil

	}
	itemPointerValue := reflect.ValueOf(itemPointer)
	if value != nil {
		itemPointerValue.Elem().Set(reflect.ValueOf(value))
	} else {
		itemPointerValue.Elem().Set(reflect.Zero(reflect.TypeOf(itemPointer).Elem()))

	}
	return nil
}

//NewSliceIterator creates a new slice iterator.
func NewSliceIterator(slice interface{}) Iterator {
	if aSlice, ok := slice.([]interface{}); ok {
		return &interfaceSliceIterator{aSlice, 0}
	}
	if aSlice, ok := slice.([]string); ok {
		return &stringSliceIterator{aSlice, 0}
	}
	sliceValue := DiscoverValueByKind(reflect.ValueOf(slice), reflect.Slice)
	return &sliceIterator{sliceValue: sliceValue}
}
