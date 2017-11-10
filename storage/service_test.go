package storage_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestStorageService_List(t *testing.T) {
	service := storage.NewService()
	assert.NotNil(t, service)
	fileName, _, _ := toolbox.CallerInfo(2)
	parent, _ := path.Split(fileName)
	baseUrl := "file://" + parent + "/test"

	if toolbox.FileExists(parent + "/test/file3.txt") {
		os.Remove(parent + "/test/file3.txt")
	}
	defer os.Remove(parent + "/test/file3.txt")

	objects, err := service.List(baseUrl)
	assert.Nil(t, err)

	assert.Equal(t, 5, len(objects))
	var objectByUrl = make(map[string]storage.Object)
	for _, object := range objects {
		objectByUrl[object.URL()] = object
	}
	assert.NotNil(t, objectByUrl[baseUrl+"/dir"])
	assert.NotNil(t, objectByUrl[baseUrl+"/file1.txt"])
	assert.NotNil(t, objectByUrl[baseUrl+"/file2.txt"])
	assert.True(t, objectByUrl[baseUrl+"/dir"].IsFolder())
	assert.True(t, objectByUrl[baseUrl+"/file2.txt"].IsContent())

	{
		reader, err := service.Download(objectByUrl[baseUrl+"/file2.txt"])
		assert.Nil(t, err)
		content, err := ioutil.ReadAll(reader)
		assert.Nil(t, err)
		assert.Equal(t, "line1\nline2", string(content))
	}

	var newFileUrl = baseUrl + "/file3.txt"
	err = service.Upload(baseUrl+"/file3.txt", bytes.NewReader([]byte("abc")))
	assert.Nil(t, err)

	exists, err := service.Exists(baseUrl + "/file3.txt")
	assert.Nil(t, err)
	assert.True(t, exists)

	{
		object, err := service.StorageObject(newFileUrl)
		assert.Nil(t, err)
		reader, err := service.Download(object)
		assert.Nil(t, err)
		content, err := ioutil.ReadAll(reader)
		assert.Nil(t, err)
		assert.Equal(t, "abc", string(content))
	}

}

