package aws

import "github.com/viant/toolbox"

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
	return result, toolbox.LoadConfigFromUrl(URL, result)
}
