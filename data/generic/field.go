package dynamic

import "reflect"

//Field represents dynamic filed
type Field struct {
	Name  string
	Type  reflect.Type
	index int
}

//Set sets a field value
func (f *Field) Set(value interface{}, result *[]interface{})  {
	values := *result
	values = reallocateIfNeeded(f.index + 1,  values)
	values[f.index] = value
	*result = values
}


//Get returns field value
func (f *Field) Get(values []interface{})  interface{} {
	if f.index <  len(values) {
		return values[f.index]
	}
	return nil
}
