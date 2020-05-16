package udf

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/data"
	"io/ioutil"
)
//LoadJSON loads new line delimited or regular JSON into data structure
func LoadJSON(source interface{}, state data.Map) (interface{}, error) {
	location := toolbox.AsString(source)
	if location == "" {
		return nil, errors.New("location was empty at LoadJSON")
	}
	data, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load: %v", location)
	}
	JSON := string(data)
	if toolbox.IsNewLineDelimitedJSON(JSON) {
		return toolbox.NewLineDelimitedJSON(JSON)
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
