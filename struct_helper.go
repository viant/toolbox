package toolbox

import (
	"fmt"
	"github.com/go-errors/errors"
	"reflect"
	"strings"
)

const (
	fieldNameKey  = "fieldName"
	anonymousKey  = "anonymous"
	fieldIndexKey = "fieldIndex"
	defaultKey    = "default"
)

var columnMapping = []string{"column", "dateLayout", "dateFormat", "autoincrement", "primaryKey", "sequence", "valueMap", defaultKey, anonymousKey}

//ScanStructFunc scan supplied struct methods
func ScanStructMethods(structOrItsType interface{}, depth int, handler func(method reflect.Method) error) error {
	var scanned = make(map[reflect.Type]bool)
	return scanStructMethods(structOrItsType, scanned, depth, handler)
}

func scanStructMethods(structOrItsType interface{}, scanned map[reflect.Type]bool, depth int, handler func(method reflect.Method) error) error {
	if depth < 0 {
		return nil
	}

	structValue, err := TryDiscoverValueByKind(reflect.ValueOf(structOrItsType), reflect.Struct)
	if err != nil {
		structValue := reflect.ValueOf(structOrItsType)
		if !(structValue.Kind() == reflect.Interface) {
			return err
		}
	}

	structType := structValue.Type()
	if _, hasScan := scanned[structType]; hasScan {
		return nil
	}

	scanned[structType] = true

	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structType.Field(i)
		if isExported := fieldType.PkgPath == ""; !isExported {
			continue
		}
		if !fieldType.Anonymous {
			continue
		}
		if !IsStruct(fieldType) {
			continue
		}
		if fieldStructType, err := TryDiscoverTypeByKind(fieldType, reflect.Struct); err == nil {
			fieldStruct := reflect.New(fieldStructType).Interface()
			if err = scanStructMethods(fieldStruct, scanned, depth-1, handler); err != nil {
				return err
			}
		}
	}

	structPtr, err := TryDiscoverValueByKind(reflect.ValueOf(structOrItsType), reflect.Ptr)
	if err != nil {
		return err
	}

	structTypePtr := structPtr.Type()
	for i := 0; i < structTypePtr.NumMethod(); i++ {
		method := structTypePtr.Method(i)
		if isExported := method.PkgPath == ""; !isExported {
			continue
		}
		if err := handler(method); err != nil {
			return err
		}
	}
	return nil
}

//StructField represents a struct field
type StructField struct {
	Owner reflect.Value
	Value reflect.Value
	Type  reflect.StructField
}

var onUnexportedHandler = IgnoreUnexportedFields

//UnexportedFieldHandler represents unexported field handler
type UnexportedFieldHandler func(structField *StructField) bool

//Handler ignoring unexported fields
func IgnoreUnexportedFields(structField *StructField) bool {
	return false
}

func SetUnexportedFieldHandler(handler UnexportedFieldHandler) error {
	if handler == nil {
		return errors.New("handler was nil")
	}
	onUnexportedHandler = handler
	return nil
}

//ProcessStruct reads passed in struct fields and values to pass it to provided handler
func ProcessStruct(aStruct interface{}, handler func(fieldType reflect.StructField, field reflect.Value) error) error {
	structValue, err := TryDiscoverValueByKind(reflect.ValueOf(aStruct), reflect.Struct)
	if err != nil {
		return err
	}
	structType := structValue.Type()

	var fields = make(map[string]*StructField)
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		if !fieldType.Anonymous {
			continue
		}
		field := structValue.Field(i)
		if !IsStruct(field) {
			fields[fieldType.Name] = &StructField{Type: fieldType, Value: field, Owner: structValue}
			continue
		}
		var aStruct interface{}
		if fieldType.Type.Kind() == reflect.Ptr {
			if field.IsNil() {
				if !field.CanSet() {
					continue
				}
				structValue.Field(i).Set(reflect.New(fieldType.Type.Elem()))
			}
			aStruct = field.Interface()
		} else {

			if field.CanAddr() {
				aStruct = field.Addr().Interface()
			} else if field.CanInterface() {
				aStruct = field.Interface()
			} else {
				continue
			}
		}
		if err := ProcessStruct(aStruct, func(fieldType reflect.StructField, field reflect.Value) error {
			structField := &StructField{Type: fieldType, Value: field, Owner: field}
			if field.CanAddr() {
				structField.Owner = field.Addr()
			}
			fields[fieldType.Name] = structField
			return nil
		}); err != nil {
			return err
		}
	}

	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		if fieldType.Anonymous {
			continue
		}
		field := structValue.Field(i)
		structField := &StructField{Owner: structValue, Type: fieldType, Value: field}
		if isExported := fieldType.PkgPath == ""; !isExported {
			if !onUnexportedHandler(structField) {
				continue
			}
		}
		fields[fieldType.Name] = &StructField{Owner: structValue, Type: fieldType, Value: field}
	}

	for _, field := range fields {
		if err := handler(field.Type, field.Value); err != nil {
			return err
		}
	}
	return nil
}

//BuildTagMapping builds map keyed by mappedKeyTag tag value, and value is another map of keys where tag name is presents in the tags parameter.
func BuildTagMapping(structTemplatePointer interface{}, mappedKeyTag string, resultExclusionTag string, inheritKeyFromField bool, convertKeyToLowerCase bool, tags []string) map[string](map[string]string) {
	reflectStructType := DiscoverTypeByKind(structTemplatePointer, reflect.Struct)
	var result = make(map[string]map[string]string)
	var anonymousMappings = make(map[string]map[string]string)

	for i := 0; i < reflectStructType.NumField(); i++ {
		var field reflect.StructField
		field = reflectStructType.Field(i)
		key := getTagValues(field, mappedKeyTag)

		if field.Anonymous && key == "" {
			var anonymousType = DereferenceType(field.Type)
			if anonymousType.Kind() == reflect.Struct {
				anonymousMapping := BuildTagMapping(reflect.New(anonymousType).Interface(), mappedKeyTag, resultExclusionTag, inheritKeyFromField, convertKeyToLowerCase, tags)
				for k, v := range anonymousMapping {
					anonymousMappings[k] = v
					anonymousMappings[k][anonymousKey] = "true"
					anonymousMappings[k][fieldIndexKey] = AsString(i)
				}
			}

			continue
		}

		isTransient := strings.EqualFold(field.Tag.Get(resultExclusionTag), "true")
		if isTransient {
			continue
		}

		if key == "" {
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
		result[key][fieldNameKey] = field.Name
	}

	for k, v := range anonymousMappings {
		if _, has := result[k]; !has {
			result[k] = v
		}
	}
	return result
}

func getTagValues(field reflect.StructField, mappedKeyTag string) string {
	key := field.Tag.Get(mappedKeyTag)
	key = strings.Split(key, ",")[0]
	if mappedKeyTag == fieldNameKey {
		key = field.Name
	}
	return key
}

//NewFieldSettingByKey reads field's tags and returns them indexed by passed in key, fieldName is always part of the resulting map unless filed has "transient" tag.
func NewFieldSettingByKey(aStruct interface{}, key string) map[string](map[string]string) {
	return BuildTagMapping(aStruct, key, "transient", true, true, columnMapping)
}

func setEmptyMap(source reflect.Value, dataTypes map[string]bool) {
	if !source.CanSet() {
		return
	}
	mapType := source.Type()

	mapPointer := reflect.New(mapType)

	mapValueType := mapType.Elem()
	mapKeyType := mapType.Key()

	newMap := mapPointer.Elem()

	newMap.Set(reflect.MakeMap(mapType))
	targetMapKeyPointer := reflect.New(mapKeyType)

	targetMapValuePointer := reflect.New(mapValueType)

	var elementKey = targetMapKeyPointer.Elem()
	var elementValue = targetMapValuePointer.Elem()

	if elementValue.Kind() == reflect.Ptr && elementValue.IsNil() {
		component := reflect.New(elementValue.Type().Elem())
		elementValue.Set(component)
	}
	if elementKey.Type() != mapKeyType {
		if elementKey.Type().AssignableTo(mapKeyType) {
			elementKey = elementKey.Convert(mapKeyType)
		}
	}

	if DereferenceType(elementValue.Type()).Kind() == reflect.Struct {
		initStruct(elementValue.Interface(), dataTypes)
	}

	newMap.SetMapIndex(elementKey, elementValue)
	var elem = mapPointer.Elem()
	source.Set(elem)
}

func createEmptySlice(source reflect.Value, dataTypes map[string]bool) {
	sliceType := DiscoverTypeByKind(source.Type(), reflect.Slice)
	if !source.CanSet() {
		return
	}
	slicePointer := reflect.New(sliceType)
	slice := slicePointer.Elem()
	componentType := DiscoverComponentType(sliceType)
	var targetComponentPointer = reflect.New(componentType)
	var targetComponent = targetComponentPointer.Elem()
	if DereferenceType(componentType).Kind() == reflect.Struct {
		componentType := targetComponent.Type()
		isPointer := componentType.Kind() == reflect.Ptr
		if isPointer {
			componentType = componentType.Elem()
		}
		structElement := reflect.New(componentType)
		initStruct(structElement.Interface(), dataTypes)

		if isPointer {
			targetComponentPointer.Elem().Set(structElement)
		} else {
			targetComponentPointer.Elem().Set(structElement.Elem())
		}
		initStruct(targetComponentPointer.Elem().Interface(), dataTypes)
	}
	slice.Set(reflect.Append(slice, targetComponentPointer.Elem()))
	source.Set(slicePointer.Elem())
}

//InitStruct initialise any struct pointer to empty struct
func InitStruct(source interface{}) {
	var dataTypes = make(map[string]bool)
	if source == nil {
		return
	}
	initStruct(source, dataTypes)
}

func initStruct(source interface{}, dataTypes map[string]bool) {
	if source == nil {
		return
	}

	if !IsStruct(source) {
		return
	}

	var key = DereferenceType(source).Name()
	if _, has := dataTypes[key]; has {
		return
	}
	dataTypes[key] = true

	sourceValue, ok := source.(reflect.Value)
	if !ok {
		sourceValue = reflect.ValueOf(source)
	}

	if sourceValue.Type().Kind() == reflect.Ptr {
		elem := sourceValue.Elem()
		if elem.Kind() == reflect.Ptr && elem.IsNil() {
			return
		}
		if !sourceValue.Elem().IsValid() {
			return
		}
	}

	_ = ProcessStruct(source, func(fieldType reflect.StructField, fieldValue reflect.Value) error {
		if !fieldValue.CanInterface() {
			return nil
		}

		if fieldValue.Kind() == reflect.String && fieldValue.CanSet() {
			fieldValue.SetString(" ")
			return nil
		}

		if fieldType.Type.Kind() == reflect.Map {
			setEmptyMap(fieldValue, dataTypes)
			return nil
		}
		if fieldType.Type.Kind() == reflect.Slice {
			createEmptySlice(fieldValue, dataTypes)
			return nil
		}
		if fieldType.Type.Kind() != reflect.Ptr {
			return nil
		}
		if DereferenceType(fieldType).Kind() == reflect.Struct {
			if !fieldValue.CanSet() {
				return nil
			}
			if fieldValue.Type().Kind() == reflect.Ptr {
				fieldStruct := reflect.New(fieldValue.Type().Elem())

				if reflect.TypeOf(source) != fieldStruct.Type() {
					initStruct(fieldStruct.Interface(), dataTypes)
				}
				fieldValue.Set(fieldStruct)
			}

		}
		return nil
	})
}

//StructFieldMeta represents struct field meta
type StructFieldMeta struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Required    bool   `json:"required,"`
	Description string `json:"description,omitempty"`
}

//StructMeta represents struct meta details
type StructMeta struct {
	Type         string
	rawType      reflect.Type       `json:"-"`
	Fields       []*StructFieldMeta `json:"fields,omitempty"`
	Dependencies []*StructMeta      `json:"dependencies,omitempty"`
}

func (m *StructMeta) Message() map[string]interface{} {
	var result = make(map[string]interface{})
	var deps = make(map[string]*StructMeta)
	for _, dep := range m.Dependencies {
		deps[dep.Type] = dep
	}
	for _, field := range m.Fields {
		if dep, ok := deps[field.Type]; ok {
			result[field.Name] = dep.Message()
			continue
		}
		result[field.Name] = ""
	}
	return result
}

//StructMetaFilter
type StructMetaFilter func(field reflect.StructField) bool

func DefaultStructMetaFilter(ield reflect.StructField) bool {
	return true
}

var structMetaFilter StructMetaFilter = DefaultStructMetaFilter

//SetStructMetaFilter sets struct meta filter
func SetStructMetaFilter(filter StructMetaFilter) error {
	if filter == nil {
		return errors.New("filter was nil")
	}
	structMetaFilter = filter
	return nil
}

//GetStructMeta returns struct meta
func GetStructMeta(source interface{}) *StructMeta {
	var result = &StructMeta{}
	var trackedTypes = make(map[string]bool)
	getStructMeta(source, result, trackedTypes)
	return result
}

//InitStruct initialise any struct pointer to empty struct
func getStructMeta(source interface{}, meta *StructMeta, trackedTypes map[string]bool) bool {
	if source == nil {
		return false
	}

	var structType = fmt.Sprintf("%T", source)
	if _, has := trackedTypes[structType]; has {
		return false
	}

	meta.Type = structType
	meta.Fields = make([]*StructFieldMeta, 0)
	meta.Dependencies = make([]*StructMeta, 0)
	sourceValue := reflect.ValueOf(source)

	if sourceValue.Kind() == reflect.Ptr {
		elem := sourceValue.Elem()
		if elem.Kind() == reflect.Ptr && elem.IsNil() {
			return false
		}

		if !sourceValue.Elem().IsValid() {
			source = reflect.New(sourceValue.Type().Elem()).Interface()
		}
	}

	meta.rawType = sourceValue.Type()
	trackedTypes[structType] = true
	_ = ProcessStruct(source, func(fieldType reflect.StructField, field reflect.Value) error {
		if !structMetaFilter(fieldType) {
			return nil
		}
		if isExported := fieldType.PkgPath == ""; !isExported {
			structField := &StructField{
				Owner: reflect.ValueOf(source),
				Type:  fieldType,
				Value: field,
			}
			if !onUnexportedHandler(structField) {
				return nil
			}
			field = structField.Value
		}

		if isJSONSkippable(string(fieldType.Tag)) {
			return nil
		}
		fieldMeta := &StructFieldMeta{}
		fieldMeta.Name = fieldType.Name
		fieldMeta.Type = fieldType.Type.Name()
		meta.Fields = append(meta.Fields, fieldMeta)

		if value, ok := fieldType.Tag.Lookup("required"); ok {
			fieldMeta.Required = AsBoolean(value)
		}
		if value, ok := fieldType.Tag.Lookup("description"); ok {
			fieldMeta.Description = value
		}
		var value = field.Interface()
		if value == nil {
			return nil
		}
		fieldMeta.Type = fmt.Sprintf("%T", value)
		if fieldType.PkgPath != "" {
			fieldMeta.Type = strings.Replace(fieldMeta.Type, "*", "", 1)
		}

		if IsStruct(value) {
			var fieldStruct = &StructMeta{}
			switch field.Kind() {
			case reflect.Ptr:
				var fieldValue interface{}
				if field.IsNil() {
					fieldValue = reflect.New(field.Type().Elem()).Interface()
				} else {
					fieldValue = field.Elem().Interface()
				}
				if getStructMeta(fieldValue, fieldStruct, trackedTypes) {
					meta.Dependencies = append(meta.Dependencies, fieldStruct)
				}

			case reflect.Struct:
				if field.CanInterface() {
					if getStructMeta(field.Interface(), fieldStruct, trackedTypes) {
						meta.Dependencies = append(meta.Dependencies, fieldStruct)
					}
				}

			}

			return nil
		}
		if IsMap(value) {
			var aMap = AsMap(field.Interface())
			var mapValue interface{}
			for _, mapValue = range aMap {
				break
			}
			if mapValue != nil && IsStruct(mapValue) {
				var fieldStruct = &StructMeta{}
				if getStructMeta(mapValue, fieldStruct, trackedTypes) {
					meta.Dependencies = append(meta.Dependencies, fieldStruct)

				}
			}
			return nil
		}
		if IsSlice(value) {
			var aSlice = AsSlice(field.Interface())
			if len(aSlice) > 0 {
				if aSlice[0] != nil && IsStruct(aSlice[0]) {
					var fieldStruct = &StructMeta{}
					if getStructMeta(aSlice[0], fieldStruct, trackedTypes) {
						meta.Dependencies = append(meta.Dependencies, fieldStruct)
					}
				}
			}
			return nil
		}
		return nil
	})
	return true
}

func isJSONSkippable(tag string) bool {
	return strings.Contains(tag, "json:\"-")
}

//StructFields by name sorter
type StructFields []*StructField

// Len is part of sort.Interface.
func (s StructFields) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s StructFields) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface.
func (s StructFields) Less(i, j int) bool {
	return s[i].Type.Name < s[j].Type.Name
}
