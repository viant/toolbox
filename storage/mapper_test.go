package storage_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestTemplateWriter_GenerateStorageCode(t *testing.T) {

	parent := toolbox.CallerDirectory(3)
	var source = toolbox.FileSchema + path.Join(parent, "test", "source")
	var destination = path.Join("test", "source")
	var target = path.Join(parent, "test", "gen", "source.go")

	parent, _ = path.Split(target)
	toolbox.CreateDirIfNotExist(parent)
	err := storage.GenerateStorageCode(&storage.StorageMapping{
		SourceURL:      source,
		DestinationURI: destination,
		TargetFile:     target,
		TargetPackage:  "gen",
	})
	assert.Nil(t, err)

	reader, err := os.Open(target)
	assert.Nil(t, err)
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if assert.Nil(t, err) {
		textContent := string(content)
		assert.Contains(t, textContent, `err := memStorage.Upload("mem://test/source/file2.txt", bytes.NewReader([]byte`)
	}

}
