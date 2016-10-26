package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"path"
	"testing"
)

func TestLoadConfigFromUrl(t *testing.T) {

	fileName, _, _ := getCallerInfo(2)
	basePath := path.Dir(fileName)
	url := toolbox.FileSchema + path.Join(basePath, "test", "config.json")
	config := &TestConfig{}
	err := toolbox.LoadConfigFromUrl(url, config)
	assert.Nil(t, err)
	assert.Equal(t, "value1", config.Attr1)
	assert.Equal(t, "value2", config.Attr2)

	//corrupted_config.json
}

func TestLoadConfigFromUrl_Corrupted(t *testing.T) {

	fileName, _, _ := getCallerInfo(2)
	basePath := path.Dir(fileName)
	url := toolbox.FileSchema + path.Join(basePath, "test", "corrupted_config.json")
	config := &TestConfig{}
	err := toolbox.LoadConfigFromUrl(url, config)
	assert.NotNil(t, err)
}

func TestLoadConfigFromUrl_NonExisting(t *testing.T) {
	fileName, _, _ := getCallerInfo(2)
	basePath := path.Dir(fileName)
	url := toolbox.FileSchema + path.Join(basePath, "test", "non_existing_config.json")
	config := &TestConfig{}
	err := toolbox.LoadConfigFromUrl(url, config)
	assert.NotNil(t, err)
}

func TestLoadConfigFromUrl_EmptyUrl(t *testing.T) {
	url := ""
	config := &TestConfig{}
	err := toolbox.LoadConfigFromUrl(url, config)
	assert.NotNil(t, err)
}

type TestConfig struct {
	Attr1 string
	Attr2 string
}
