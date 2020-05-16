package udf

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

//Length returns length of slice or string
func Length(source interface{}, state data.Map) (interface{}, error) {

	if toolbox.IsSlice(source) {
		return len(toolbox.AsSlice(source)), nil
	}
	if toolbox.IsMap(source) {
		return len(toolbox.AsMap(source)), nil
	}

	if text, ok := source.(string); ok {
		if strings.HasPrefix(text, "$") {
			return nil, fmt.Errorf("unexpanded variable: %v", text)
		}
		return len(text), nil
	}
	return 0, nil
}


//Replace replaces text with old and new fragments
func Replace(source interface{}, state data.Map) (interface{}, error) {
	var args []interface{}
	if ! toolbox.IsSlice(source) {
		return nil, fmt.Errorf("expacted %T, but had %T", args, source)
	}
	args = toolbox.AsSlice(source)
	if len(args) < 3 {
		return nil, fmt.Errorf("expected 3 arguments (text, old, new), but had: %v" , len(args))
	}
	text := toolbox.AsString(args[0])
	old := toolbox.AsString(args[1])
	new := toolbox.AsString(args[2])
	count := strings.Count(text, old)
	return strings.Replace(text, old, new, count), nil
}


// Join joins slice by separator
func Join(args interface{}, state data.Map) (interface{}, error) {
	if !toolbox.IsSlice(args) {
		return nil, fmt.Errorf("expected 2 arguments but had: %T", args)
	}
	arguments := toolbox.AsSlice(args)
	if len(arguments) != 2 {
		return nil, fmt.Errorf("expected 2 arguments but had: %v", len(arguments))
	}

	if !toolbox.IsSlice(arguments[0]) {
		return nil, fmt.Errorf("expected 1st arguments as slice but had: %T", arguments[0])
	}
	var result = make([]string, 0)
	toolbox.CopySliceElements(arguments[0], &result)
	return strings.Join(result, toolbox.AsString(arguments[1])), nil
}

// Split split text to build a slice
func Split(args interface{}, state data.Map) (interface{}, error) {
	if !toolbox.IsSlice(args) {
		return nil, fmt.Errorf("expected 2 arguments but had: %T", args)
	}
	arguments := toolbox.AsSlice(args)
	if len(arguments) != 2 {
		return nil, fmt.Errorf("expected 2 arguments but had: %v", len(arguments))
	}
	if !toolbox.IsString(arguments[0]) {
		return nil, fmt.Errorf("expected 1st arguments as string but had: %T", arguments[0])
	}
	result := strings.Split(toolbox.AsString(arguments[0]), toolbox.AsString(arguments[1]))
	for i := range result {
		result[i] = strings.TrimSpace(result[i])
	}
	return result, nil
}

//Keys returns keys of the supplied map
func Keys(source interface{}, state data.Map) (interface{}, error) {
	aMap, err := AsMap(source, state)
	if err != nil {
		return nil, err
	}
	var result = make([]interface{}, 0)
	err = toolbox.ProcessMap(aMap, func(key, value interface{}) bool {
		result = append(result, key)
		return true
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

//Values returns values of the supplied map
func Values(source interface{}, state data.Map) (interface{}, error) {
	aMap, err := AsMap(source, state)
	if err != nil {
		return nil, err
	}
	var result = make([]interface{}, 0)
	err = toolbox.ProcessMap(aMap, func(key, value interface{}) bool {
		result = append(result, value)
		return true
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

//IndexOf returns index of the matched slice elements or -1
func IndexOf(source interface{}, state data.Map) (interface{}, error) {
	if !toolbox.IsSlice(source) {
		return nil, fmt.Errorf("expected arguments but had: %T", source)
	}
	args := toolbox.AsSlice(source)
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments but had: %v", len(args))
	}

	if toolbox.IsString(args[0]) {
		return strings.Index(toolbox.AsString(args[0]), toolbox.AsString(args[1])), nil
	}
	collection, err := AsCollection(args[0], state)
	if err != nil {
		return nil, err
	}
	for i, candidate := range toolbox.AsSlice(collection) {
		if candidate == args[1] || toolbox.AsString(candidate) == toolbox.AsString(args[1]) {
			return i, nil
		}
	}
	return -1, nil
}

//Base64Decode encodes source using base64.StdEncoding
func Base64Encode(source interface{}, state data.Map) (interface{}, error) {
	if source == nil {
		return "", nil
	}
	switch value := source.(type) {
	case string:
		return base64.StdEncoding.EncodeToString([]byte(value)), nil
	case []byte:
		return base64.StdEncoding.EncodeToString(value), nil
	default:
		if toolbox.IsMap(source) || toolbox.IsSlice(source) {
			encoded, err := json.Marshal(source)
			fmt.Printf("%s %v\n", encoded, err)
			if err == nil {
				return base64.StdEncoding.EncodeToString(encoded), nil
			}
		}
		return nil, fmt.Errorf("unsupported type: %T", source)
	}
}

//Base64Decode decodes source using base64.StdEncoding
func Base64Decode(source interface{}, state data.Map) (interface{}, error) {
	if source == nil {
		return "", nil
	}
	switch value := source.(type) {
	case string:
		return base64.StdEncoding.DecodeString(value)
	case []byte:
		return base64.StdEncoding.DecodeString(string(value))
	default:
		return nil, fmt.Errorf("unsupported type: %T", source)
	}
}

//Base64DecodeText decodes source using base64.StdEncoding to string
func Base64DecodeText(source interface{}, state data.Map) (interface{}, error) {
	decoded, err := Base64Decode(source, state)
	if err != nil {
		return nil, err
	}
	return toolbox.AsString(decoded), nil
}

//QueryEscape returns url escaped text
func QueryEscape(source interface{}, state data.Map) (interface{}, error) {
	text := toolbox.AsString(source)
	return url.QueryEscape(text), nil
}

//QueryUnescape returns url escaped text
func QueryUnescape(source interface{}, state data.Map) (interface{}, error) {
	text := toolbox.AsString(source)
	return url.QueryUnescape(text)
}

//TrimSpace returns trims spaces from supplied text
func TrimSpace(source interface{}, state data.Map) (interface{}, error) {
	text := toolbox.AsString(source)
	return strings.TrimSpace(text), nil
}

//Count returns count of matched nodes leaf value
func Count(xPath interface{}, state data.Map) (interface{}, error) {
	result, err := aggregate(xPath, state, func(previous, newValue float64) float64 {
		return previous + 1
	})
	if err != nil {
		return nil, err
	}
	return AsNumber(result, nil)
}

//Sum returns sums of matched nodes leaf value
func Sum(xPath interface{}, state data.Map) (interface{}, error) {
	result, err := aggregate(xPath, state, func(previous, newValue float64) float64 {
		return previous + newValue
	})
	if err != nil {
		return nil, err
	}
	return AsNumber(result, nil)
}

//Select returns all matched attributes from matched nodes, attributes can be alised with sourcePath:alias
func Select(params interface{}, state data.Map) (interface{}, error) {
	var arguments = make([]interface{}, 0)
	if toolbox.IsSlice(params) {
		arguments = toolbox.AsSlice(params)
	} else {
		arguments = append(arguments, params)
	}
	xPath := toolbox.AsString(arguments[0])
	var result = make([]interface{}, 0)
	attributes := make([]string, 0)
	for i := 1; i < len(arguments); i++ {
		attributes = append(attributes, toolbox.AsString(arguments[i]))
	}
	err := matchPath(xPath, state, func(matched interface{}) error {
		if len(attributes) == 0 {
			result = append(result, matched)
			return nil
		}
		if !toolbox.IsMap(matched) {
			return fmt.Errorf("expected map for %v, but had %T", xPath, matched)
		}
		matchedMap := data.Map(toolbox.AsMap(matched))
		var attributeValues = make(map[string]interface{})
		for _, attr := range attributes {
			if strings.Contains(attr, ":") {
				kvPair := strings.SplitN(attr, ":", 2)
				value, has := matchedMap.GetValue(kvPair[0])
				if !has {
					continue
				}
				attributeValues[kvPair[1]] = value
			} else {
				value, has := matchedMap.GetValue(attr)
				if !has {
					continue
				}
				attributeValues[attr] = value
			}
		}
		result = append(result, attributeValues)
		return nil
	})
	return result, err
}

//AsNumber return int or float
func AsNumber(value interface{}, state data.Map) (interface{}, error) {
	floatValue := toolbox.AsFloat(value)
	if float64(int(floatValue)) == floatValue {
		return int(floatValue), nil
	}
	return floatValue, nil
}

//Aggregate applies an aggregation function to matched path
func aggregate(xPath interface{}, state data.Map, agg func(previous, newValue float64) float64) (float64, error) {
	var result = 0.0
	if state == nil {
		return 0.0, fmt.Errorf("state was empty")
	}
	err := matchPath(toolbox.AsString(xPath), state, func(value interface{}) error {
		if value == nil {
			return nil
		}
		floatValue, err := toolbox.ToFloat(value)
		if err != nil {
			return err
		}
		result = agg(result, floatValue)
		return nil
	})
	return result, err
}

func matchPath(xPath string, state data.Map, handler func(value interface{}) error) error {
	fragments := strings.Split(toolbox.AsString(xPath), "/")
	var node = state
	var nodeValue interface{}
	for i, part := range fragments {

		isLast := i == len(fragments)-1
		if isLast {
			if part == "*" {
				if toolbox.IsSlice(nodeValue) {
					for _, item := range toolbox.AsSlice(nodeValue) {
						if err := handler(item); err != nil {
							return err
						}
					}
					return nil
				} else if toolbox.IsMap(nodeValue) {
					for _, item := range toolbox.AsMap(nodeValue) {
						if err := handler(item); err != nil {
							return err
						}
					}
				}
				return handler(nodeValue)
			}

			if !node.Has(part) {
				break
			}
			if err := handler(node.Get(part)); err != nil {
				return err
			}
			continue
		}
		if part != "*" {
			nodeValue = node.Get(part)
			if nodeValue == nil {
				break
			}
			if toolbox.IsMap(nodeValue) {
				node = toolbox.AsMap(nodeValue)
				continue
			}
			if toolbox.IsSlice(nodeValue) {
				continue
			}
			break
		}

		if nodeValue == nil {
			break
		}
		subXPath := strings.Join(fragments[i+1:], "/")
		if toolbox.IsSlice(nodeValue) {
			aSlice := toolbox.AsSlice(nodeValue)
			for _, item := range aSlice {
				if toolbox.IsMap(item) {
					if err := matchPath(subXPath, toolbox.AsMap(item), handler); err != nil {
						return err
					}
					continue
				}
				return fmt.Errorf("unsupported path type:%T", item)
			}
		}
		if toolbox.IsMap(nodeValue) {
			aMap := toolbox.AsMap(nodeValue)
			for _, item := range aMap {
				if toolbox.IsMap(item) {
					if err := matchPath(subXPath, toolbox.AsMap(item), handler); err != nil {
						return err
					}
					continue
				}
				return fmt.Errorf("unsupported path type:%T", item)
			}
		}
		break
	}
	return nil
}

//Rand returns random
func Rand(params interface{}, state data.Map) (interface{}, error) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	floatValue := generator.Float64()
	if params == nil || !toolbox.IsSlice(params) {
		return floatValue, nil
	}
	parameters := toolbox.AsSlice(params)
	if len(parameters) != 2 {
		return floatValue, nil
	}
	min := toolbox.AsInt(parameters[0])
	max := toolbox.AsInt(parameters[1])
	return min + int(float64(max-min)*floatValue), nil
}

//Concat concatenate supplied parameters, parameters
func Concat(params interface{}, state data.Map) (interface{}, error) {
	if params == nil || !toolbox.IsSlice(params) {
		return nil, fmt.Errorf("invalid signature, expected: $Concat(arrayOrItem1, arrayOrItem2)")
	}
	var result = make([]interface{}, 0)
	parameters := toolbox.AsSlice(params)
	if len(parameters) == 0 {
		return result, nil
	}

	if toolbox.IsString(parameters[0]) {
		result := ""
		for _, item := range parameters {
			result += toolbox.AsString(item)
		}
		return result, nil
	}

	for _, item := range parameters {
		if toolbox.IsSlice(item) {
			itemSlice := toolbox.AsSlice(item)
			result = append(result, itemSlice...)
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

//Merge creates a new merged map for supplied maps,  (mapOrPath1, mapOrPath2, mapOrPathN)
func Merge(params interface{}, state data.Map) (interface{}, error) {
	if params == nil || !toolbox.IsSlice(params) {
		return nil, fmt.Errorf("invalid signature, expected: $Merge(map1, map2, override)")
	}
	var result = make(map[string]interface{})
	parameters := toolbox.AsSlice(params)
	if len(parameters) == 0 {
		return result, nil
	}
	var ok bool
	for _, item := range parameters {
		if toolbox.IsString(item) && state != nil {
			if item, ok = state.GetValue(toolbox.AsString(item)); !ok {
				continue
			}
		}
		if !toolbox.IsMap(item) {
			continue
		}
		itemMap := toolbox.AsMap(item)
		for k, v := range itemMap {
			result[k] = v
		}
	}
	return result, nil
}

//AsNewLineDelimitedJSON convers a slice into new line delimited JSON
func AsNewLineDelimitedJSON(source interface{}, state data.Map) (interface{}, error) {
	if source == nil || !toolbox.IsSlice(source) {
		return nil, fmt.Errorf("invalid signature, expected: $AsNewLineDelimitedJSON([])")
	}
	aSlice := toolbox.AsSlice(source)
	var result = make([]string, 0)
	for _, item := range aSlice {
		data, _ := json.Marshal(item)
		result = append(result, string(data))
	}
	return strings.Join(result, "\n"), nil
}