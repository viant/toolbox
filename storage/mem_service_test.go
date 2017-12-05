package storage_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/storage"
	"io/ioutil"
	"strings"
	"testing"
)

func Test_NewMemoryService(t *testing.T) {
	service := storage.NewMemoryService()

	var files = map[string]string{
		"mem:///test/file1.txt":     "abc",
		"mem:///test/file2.txt":     "xyz",
		"mem:///test/sub/file1.txt": "---",
		"mem:///test/sub/file2.txt": "xxx",
	}

	for k, v := range files {
		err := service.Upload(k, strings.NewReader(v))
		assert.Nil(t, err)
	}

	for k, v := range files {
		object, err := service.StorageObject(k)
		if assert.Nil(t, err) {
			reader, err := service.Download(object)
			if assert.Nil(t, err) {
				content, err := ioutil.ReadAll(reader)
				assert.Nil(t, err)
				assert.Equal(t, v, string(content))

			}
		}
	}

	objects, err := service.List("mem:///test/sub/")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(objects))

	assert.True(t, objects[0].IsFolder())
	for k := range files {
		object, err := service.StorageObject(k)
		if assert.Nil(t, err) {
			err = service.Delete(object)
			assert.Nil(t, err)
		}
	}
}
