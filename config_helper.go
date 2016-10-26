package toolbox

import (
	"fmt"
	"github.com/viant/jdsunit/target/src/github.com/viant/toolbox"
	"reflect"
)

//LoadConfigFromUrl loads json configuration from url to passed in config pointer.
func LoadConfigFromUrl(url string, config interface{}) error {
	var configType = reflect.TypeOf(config).Elem().Name()
	if len(url) == 0 {
		return fmt.Errorf("%v and %Url were empty", configType, configType)
	}
	reader, _, err := toolbox.OpenReaderFromURL(url)
	if err != nil {
		return fmt.Errorf("Failed to load %v from url %v %v", configType, url, err)
	}

	err = toolbox.NewJSONDecoderFactory().Create(reader).Decode(config)
	if err != nil {
		return fmt.Errorf("Failed to decode DbConfig from url %v %v", url, err)
	}
	return nil
}
