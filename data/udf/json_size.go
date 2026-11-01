package udf

import (
	"encoding/json"
	"github.com/viant/toolbox/data"
)

// JSONSize returns the UTF-8 byte length of the JSON representation, including
// escaping. Unlike Len, it measures an encoded request rather than its members.
func JSONSize(source interface{}, _ data.Map) (interface{}, error) {
	encoded, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	return len(encoded), nil
}
