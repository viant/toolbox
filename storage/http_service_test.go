package storage_test

import (
	"testing"
	"github.com/viant/toolbox/storage"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
)


func TestNewHttpStorageService(t *testing.T) {
	credentialFile := ""
	sourceService, err := storage.NewServiceForURL("https://github.com/viant/", credentialFile)
	assert.Nil(t, err)
	assert.NotNil(t, sourceService)
	objects, err := sourceService.List("https://github.com/viant/")
	assert.True(t, len(objects) > 0)

	reader, err := sourceService.Download(objects[0])
	assert.Nil(t, err)
	content, err := ioutil.ReadAll(reader)
	assert.Nil(t, err)
	assert.True(t, len(content) > 0)

}