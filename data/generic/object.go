package dynamic

import (
	"encoding/json"
)

//Object represents dynamic object
type Object struct {
	meta  *Meta
	_data []interface{}
}

///Sets object value
func (o *Object) Set(values map[string]interface{}) {
	o._data = o.meta.asValues(values)
}

//AsMap return map
func (o *Object) AsMap() map[string]interface{} {
	return o.meta.asMap(o._data)
}

//SetValue sets values
func (o *Object) SetValue(name string, value interface{}) {
	field := o.meta.getField(name, value)
	field.Set(value, &o._data)
}

//GetValue get values
func (o *Object) GetValue(name string) interface{} {
	field := o.meta.Field(name)
	if field == nil {
		return nil
	}
	return field.Get(o._data)
}

//MarshalJSON converts object to JSON object
func (d Object) MarshalJSON() ([]byte, error) {
	aMap := d.meta.asMap(d._data)
	return json.Marshal(aMap)
}
