package udf

import (
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
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
	if toolbox.IsSlice(source) {
		return nil, fmt.Errorf("expected arguments but had: %T", source)
	}
	args := toolbox.AsSlice(source)
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments but had: %v", len(args))
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
