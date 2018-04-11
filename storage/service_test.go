package storage_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	_ "github.com/viant/toolbox/storage/scp"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
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

	assert.True(t, len(objects) >= 5)
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
		if assert.Nil(t, err) {
			defer reader.Close()
			content, err := ioutil.ReadAll(reader)
			assert.Nil(t, err)
			assert.Equal(t, "line1\nline2", string(content))
		}
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
		if assert.Nil(t, err) {
			defer reader.Close()
			content, err := ioutil.ReadAll(reader)
			assert.Nil(t, err)
			assert.Equal(t, "abc", string(content))
		}
	}

}

func TestUpload(t *testing.T) {

	var path = "/tmp/local/test.txt"
	toolbox.RemoveFileIfExist(path)
	exec.Command("rmdir /tmp/local").CombinedOutput()
	var destination = "scp://127.0.0.1/" + path

	service, err := storage.NewServiceForURL(destination, "")
	assert.Nil(t, err)

	err = service.Upload(destination, strings.NewReader("abc"))
	assert.Nil(t, err)

}
