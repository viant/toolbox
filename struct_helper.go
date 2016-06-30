package toolbox

import (
	"reflect"
	"strings"
)

var columnMapping = []string{"column", "dateLayout", "dateFormat", "autoincrement", "primaryKey", "sequence", "valueMap"}

//ProcessStruct reads passed in struct fields and values to pass it to provided handler
func ProcessStruct(aStruct interface{}, handler func(field reflect.StructField, value interface{})) {
	structValue := DiscoverValueByKind(reflect.ValueOf(aStruct), reflect.Struct)
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		fieldStruct := structType.Field(i)
		fieldName := fieldStruct.Name
		if strings.ToLower(fieldName[0:1]) == fieldName[0:1] {
			//skip private fileds
			continue
		}
		field := structValue.Field(i)
		value := UnwrapValue(&field)
		handler(fieldStruct, value)
	}
}

//BuildTagMapping builds map keyed by mappedKeyTag tag value, and value is another map of keys where tag name is presents in the tags parameter.
func BuildTagMapping(structTemplatePointer interface{}, mappedKeyTag string, resultExclusionTag string, inheritKeyFromField bool, convertKeyToLowerCase bool, tags []string) map[string](map[string]string) {
	reflectStructType := DiscoverTypeByKind(structTemplatePointer, reflect.Struct)
	var result = make(map[string]map[string]string)
	for i := 0; i < reflectStructType.NumField(); i++ {
		var field reflect.StructField
		field = reflectStructType.Field(i)
		isTransient := strings.EqualFold(field.Tag.Get(resultExclusionTag), "true")
		if isTransient {
			continue
		}

		key := field.Tag.Get(mappedKeyTag)
		if mappedKeyTag == "fieldName" {
			key = field.Name
		}
		if len(key) == 0 {
			if !inheritKeyFromField {
				continue
			}
			key = field.Name
		}

		if convertKeyToLowerCase {
			key = strings.ToLower(key)
		}

		result[key] = make(map[string]string)
		for _, tag := range tags {
			tagValue := field.Tag.Get(tag)
			if len(tagValue) > 0 {
				result[key][tag] = tagValue
			}
		}
		result[key]["fieldName"] = field.Name
	}
	return result
}

//NewFieldSettingByKey reads field's tags and returns them indexed by passed in key, fieldName is always part of the resulting map unless filed has "transient" tag.
func NewFieldSettingByKey(aStruct interface{}, key string) map[string](map[string]string) {
	return BuildTagMapping(aStruct, key, "transient", true, true, columnMapping)
}
