package storage_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	_ "github.com/viant/toolbox/storage/scp"
	"path"
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
	//
	//{ //copy file to dir
	//
	//	sourceURL := toolbox.URLPathJoin(baseUrl, "source/dir/file.json")
	//	targetURL := toolbox.URLPathJoin(baseUrl, "target/dir3/")
	//
	//	err := storage.Copy(service, sourceURL, service, targetURL, nil, nil)
	//	assert.Nil(t, err)
	//
	//	expectedFiles := []string{
	//		path.Join(parent, "test/target/dir3/file.json"),
	//	}
	//	for _, file := range expectedFiles {
	//		assert.True(t, toolbox.FileExists(file))
	//		os.Remove(file)
	//	}
	//}
	//{ //copy file to file
	//
	//	sourceURL := path.Join(baseUrl, "source/dir/file.json")
	//	targetURL := path.Join(baseUrl, "target/dir4/file.json")
	//
	//	err := storage.Copy(service, sourceURL, service, targetURL, nil, nil)
	//	assert.Nil(t, err)
	//
	//	expectedFiles := []string{
	//		path.Join(parent, "test/target/dir4/file.json"),
	//	}
	//	for _, file := range expectedFiles {
	//		assert.True(t, toolbox.FileExists(file))
	//		os.Remove(file)
	//	}
	//}
}
