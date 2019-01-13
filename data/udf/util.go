package udf

import (
	"encoding/base64"
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"net/url"
	"strings"
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
		return len(text), nil
	}
	return 0, nil
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
	if ! toolbox.IsSlice(source) {
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
	if err !=nil {
		return nil, err
	}
	return toolbox.AsString(decoded), nil
}


//QueryEscape returns url escaped text
func QueryEscape(source interface{}, state data.Map) (interface{}, error) {
	text := toolbox.AsString(source)
	return url.QueryEscape(text), nil
}
