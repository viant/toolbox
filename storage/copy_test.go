package storage_test

import (
	"archive/zip"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	_ "github.com/viant/toolbox/storage/scp"
	"os"
	"path"
	"strings"
	"testing"
)

func TestCopy(t *testing.T) {
	service := storage.NewService()
	assert.NotNil(t, service)

	parent := toolbox.CallerDirectory(3)
	baseUrl := "file://" + parent + "/test"

	toolbox.CreateDirIfNotExist(path.Join(parent, "test/target"))

	{
		sourceURL := path.Join(baseUrl, "source/")
		targetURL := path.Join(baseUrl, "target/")
		err := storage.Copy(service, sourceURL, service, targetURL, nil, nil)
		assert.Nil(t, err)

		expectedFiles := []string{
			path.Join(parent, "test/target/file1.txt"),
			path.Join(parent, "test/target/file2.txt"),
			path.Join(parent, "test/target/dir/file.json"),
			path.Join(parent, "test/target/dir2/subdir/file1.txt"),
		}
		for _, file := range expectedFiles {
			assert.True(t, toolbox.FileExists(file))
			//os.Remove(file)
		}
	}
}

func TestArchive(t *testing.T) {
	memService := storage.NewMemoryService()
	memService.Upload("mem://test/copy/archive/file1.txt", strings.NewReader("abc"))
	memService.Upload("mem://test/copy/archive/file2.txt", strings.NewReader("xyz"))
	memService.Upload("mem://test/copy/archive/config/test.prop", strings.NewReader("123"))
	toolbox.RemoveFileIfExist("/tmp/testCopy.zip")
	var writer, err = os.OpenFile("/tmp/testCopy.zip", os.O_CREATE|os.O_WRONLY, 06444)
	if assert.Nil(t, err) {
		defer writer.Close()
		archive := zip.NewWriter(writer)
		err = storage.Archive(memService, "mem://test/copy/archive/", archive)
		assert.Nil(t, err)
		archive.Flush()
		archive.Close()
	}
}
