package toolbox

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
)

//AsYamlText converts data structure int text YAML
func AsYamlText(source interface{}) (string, error) {
	if IsStruct(source) || IsMap(source) || IsSlice(source) {
		buf := new(bytes.Buffer)
		err := yaml.NewEncoder(buf).Encode(source)
		return buf.String(), err
	}
	return "", fmt.Errorf("unsupported type: %T", source)
}

//NormalizeKVPairs converts slice of KV paris into a map, and map[interface{}]interface{} to map[string]interface{}
func NormalizeKVPairs(source interface{}) (interface{}, error) {
	isDataStruct := IsMap(source) || IsStruct(source) ||  IsSlice(source)
	var err error
	var normalized interface{}
	if isDataStruct {
		var aMap = make(map[string]interface{})
		if err = ProcessMap(source, func(k, value interface{}) bool {
			var key = AsString(k)
			aMap[key] = value
			if IsMap(value) || IsSlice(value) || IsStruct(value) {
					if normalized, err = NormalizeKVPairs(value);err == nil {
						aMap[key] = normalized
					}
			}
			return true
		});err == nil {
			return aMap, nil
		}
		if IsSlice(source) { //yaml style map conversion if applicable
			aSlice := AsSlice(source)
			if len(aSlice) == 0 {
				return source, nil
			}
			if IsMap(aSlice[0]) || IsStruct(aSlice[0]) {
				if item, err := NormalizeKVPairs(aSlice[0]);err == nil {
					return []interface{}{item}, nil
				}

			} else if IsSlice(aSlice[0]) {
				for i, item := range aSlice {
					if normalized, err = NormalizeKVPairs(item);err == nil {
						aSlice[i] = normalized
					} else {
						return nil, err
					}
				}
			}
			return aSlice, nil
		}
	}
	return source, err
}
