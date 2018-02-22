package aws

import (
	"github.com/viant/toolbox/url"
)

//Config represents storage
type Config struct {
	Region string
	Key    string
	Secret string
	Token  string
}

//NewConfig creates a new config from URL
func NewConfig(URL string) (*Config, error) {
	var result = &Config{}
	resource := url.NewResource(URL)
	return result, resource.JSONDecode(result)
}
