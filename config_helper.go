package toolbox

import (
	"fmt"
	"reflect"
)

//LoadConfigFromUrl loads json configuration from url to passed in config pointer.
func LoadConfigFromUrl(url string, config interface{}) error {
	var configType = reflect.TypeOf(config).Elem().Name()
	if len(url) == 0 {
		return fmt.Errorf("%v and %vUrl were empty", configType, configType)
	}
	reader, _, err := OpenReaderFromURL(url)
	if err != nil {
		return fmt.Errorf("Failed to load %v from url %v %v", configType, url, err)
	}
	reader.Close()
	err = NewJSONDecoderFactory().Create(reader).Decode(config)
	if err != nil {
		return fmt.Errorf("Failed to decode Config from url %v %v", url, err)
	}
	return nil
}
