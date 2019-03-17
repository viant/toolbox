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
	if source == nil {
		return source, nil
	}
	isDataStruct := IsMap(source) || IsStruct(source) || IsSlice(source)
	var err error
	var normalized interface{}
	if isDataStruct {
		var aMap = make(map[string]interface{})

		err = ProcessMap(source, func(k, value interface{}) bool {
			var key = AsString(k)
			aMap[key] = value
			if value == nil {
				return true
			}
			if IsMap(value) || IsSlice(value) || IsStruct(value) {
				if normalized, err = NormalizeKVPairs(value); err == nil {
					aMap[key] = normalized
				}
			}
			return true
		})
		if err == nil {
			return aMap, nil
		}
		if IsSlice(aMap) {
			return source, err
		}
		if IsSlice(source) { //yaml style map conversion if applicable
			aSlice := AsSlice(source)
			if len(aSlice) == 0 {
				return source, nil
			}
			for i, item := range aSlice {
				if item == nil {
					continue
				}
				if IsMap(item) || IsSlice(item) {
					if normalized, err = NormalizeKVPairs(item); err == nil {
						aSlice[i] = normalized
					} else {
						return source, nil
					}
				}
			}
			return aSlice, nil
		}
	}
	return source, err
}
