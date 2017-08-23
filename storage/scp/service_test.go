package scp_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/viant/toolbox/storage/scp"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func TestService_List(t *testing.T) {

	service := scp.NewService(nil)

	assert.NotNil(t, service)

	dir, home := path.Split(os.Getenv("HOME"))
	objects, err := service.List("scp://127.0.0.1/" + dir)
	assert.Nil(t, err)
	for _, object := range objects {
		if strings.HasSuffix(object.URL(), home) {
			assert.True(t, object.IsFolder())
		}
	}

}

func TestService_Delete(t *testing.T) {

	service := scp.NewService(nil)

	assert.NotNil(t, service)

	err := service.Upload("scp://127.0.0.1//tmp/file.txt", strings.NewReader("this is test"))
	assert.Nil(t, err)

	objects, err := service.List("scp://127.0.0.1/tmp/")
	assert.Nil(t, err)
	assert.True(t, len(objects) > 1)

	object, err := service.StorageObject("scp://127.0.0.1//tmp/file.txt")
	assert.Nil(t, err)

	reader, err := service.Download(object)
	if assert.Nil(t, err) {
		content, err := ioutil.ReadAll(reader)
		assert.Nil(t, err)
		assert.Equal(t, "this is test", string(content))
	}

	err = service.Delete(object)
	assert.Nil(t, err)

}
