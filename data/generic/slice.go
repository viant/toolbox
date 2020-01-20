package dynamic

import (
	"bytes"
	"encoding/json"
)

//Slice represents dynamic slice
type Slice struct {
	meta  *Meta
	_data [][]interface{}
}

//Add add elements to a slice
func (s *Slice) Add(aMap map[string]interface{}) {
	values := s.meta.asValues(aMap)
	data := s._data
	data = append(data, values)
	s._data = data
}

//Range iterate over slice
func (s *Slice) Range(handler func(item interface{}) (bool, error)) error {
	data := s._data
	for _, item := range data {
		aMap := s.meta.asMap(item)
		if next, err := handler(aMap); !next || err != nil {
			return err
		}
	}
	return nil
}

//Range iterate over slice of object, update to objects are applied to the slice
func (s *Slice) Objects(handler func(item *Object) (bool, error)) error {
	data := s._data
	object := &Object{meta:s.meta}
	for i, item := range data {
		object._data = item
		next, err := handler(object)
		data[i] = object._data
		if !next || err != nil {
			return err
		}
	}
	return nil
}



//MarshalJSON converts slice item to JSON array.
func (d Slice) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.Write([]byte("["))
	if err != nil {
		return nil, err
	}
	i := 0
	if err = d.Range(func(item interface{}) (b bool, err error) {
		if i > 0 {
			_, err := buf.Write([]byte(","))
			if err != nil {
				return false, err
			}
		}
		i++
		data, err :=json.Marshal(item)
		if err != nil {
			return false, err
		}
		_, err = buf.Write(data)
		return err == nil, err
	});err != nil {
		return nil, err
	}
	if _, err := buf.Write([]byte("]")); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}



