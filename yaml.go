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
