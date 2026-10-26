package data

import (
	"reflect"
	"strings"
)

// structPathFields exposes one typed node without expanding strings, calling
// methods, serializing nested payloads, or mutating the caller's value.
func structPathFields(source interface{}) (Map, bool) {
	value := reflect.ValueOf(source)
	for value.IsValid() && (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return nil, false
		}
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return nil, false
	}
	fields := NewMap()
	for _, field := range reflect.VisibleFields(value.Type()) {
		if field.PkgPath != "" {
			continue
		}
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "-" {
			continue
		}
		current, available := value, true
		for _, index := range field.Index {
			for current.Kind() == reflect.Ptr || current.Kind() == reflect.Interface {
				if current.IsNil() {
					available = false
					break
				}
				current = current.Elem()
			}
			if !available {
				break
			}
			current = current.Field(index)
		}
		if !available || !current.CanInterface() {
			continue
		}
		fields[field.Name] = current.Interface()
		if jsonName != "" {
			fields[jsonName] = current.Interface()
		}
	}
	return fields, true
}
