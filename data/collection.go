package data

import (
	"github.com/viant/toolbox"
	"strings"
)

//Collection represents a slice of interface{} (generic type)
type Collection []interface{}

//Push appends provided value to the slice
func (s *Collection) Push(value interface{}) {
	(*s) = append(*s, value)
}

//PadWithMap creates missing elements with a map
func (s *Collection) PadWithMap(size int) {
	for i := len(*s); i < size; i++ {
		s.Push(NewMap())
	}
}

//Range iterates over every item in this collection as long as handler returns true. Handler takes  an index and index of the slice element.
func (s *Collection) Range(handler func(item interface{}, index int) (bool, error)) error {
	for i, elem := range *s {
		next, err := handler(elem, i)
		if err != nil {
			return err
		}
		if !next {
			break
		}

	}
	return nil
}

//RangeMap iterates every map item in this collection as long as handler returns true. Handler takes  an index and index of the slice element
func (s *Collection) RangeMap(handler func(item Map, index int) (bool, error)) error {
	var next bool
	var err error
	for i, elem := range *s {
		var aMap, ok = elem.(Map)
		if !ok {
			next, err = handler(nil, i)
		} else {
			next, err = handler(aMap, i)
		}
		if err != nil {
			return err
		}
		if !next {
			break
		}

	}
	return nil
}

//RangeMap iterates every string item in this collection as long as handler returns true. Handler takes  an index and index of the slice element
func (s *Collection) RangeString(handler func(item interface{}, index int) (bool, error)) error {
	for i, elem := range *s {
		next, err := handler(toolbox.AsString(elem), i)
		if err != nil {
			return err
		}
		if !next {
			break
		}

	}
	return nil

}

//RangeMap iterates every int item in this collection as long as handler returns true. Handler takes  an index and index of the slice element
func (s *Collection) RangeInt(handler func(item interface{}, index int) (bool, error)) error {
	for i, elem := range *s {
		next, err := handler(toolbox.AsInt(elem), i)
		if err != nil {
			return err
		}
		if !next {
			break
		}

	}
	return nil
}

//String returns a string representation of this collection
func (s *Collection) String() string {
	var items = make([]string, 0)
	for _, item := range *s {
		items = append(items, toolbox.AsString(item))
	}
	return "[" + strings.Join(items, ",") + "]"
}

//NewCollection creates a new collection and returns a pointer
func NewCollection() *Collection {
	var result Collection = make([]interface{}, 0)
	return &result
}
