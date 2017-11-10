package storage_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"path"
	"os"
	"testing"
	"github.com/viant/toolbox/storage"
	_ 	"github.com/viant/toolbox/storage/scp"
	"os/exec"
	"fmt"
)


func TestCopy(t *testing.T) {
	service := storage.NewService()
	assert.NotNil(t, service)

	fileName, _, _ := toolbox.CallerInfo(2)
	parent, _ := path.Split(fileName)
	baseUrl := "file://" + parent + "/test"


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
		os.Remove(file)
	}
}



func TestScpCopy(t *testing.T) {
	var credential = path.Join(os.Getenv("HOME"), "secret/scp.json")
	if !toolbox.FileExists(credential) {
		return
	}
	fileName, _, _ := toolbox.CallerInfo(2)
	parent, _ := path.Split(fileName)

	var destinationPath = fmt.Sprintf("%vtest/target", parent)
	_ , err := exec.Command("rm", "-rf", destinationPath).CombinedOutput()
	assert.Nil(t, err)

	baseUrl := "scp://127.0.0.1" + parent
	sourceURL := toolbox.URLPathJoin(baseUrl, "test/source/")
	targetURL := toolbox.URLPathJoin(baseUrl, "test/target/")

	service, err := storage.NewServiceForURL(sourceURL, credential)
	if assert.Nil(t, err) {
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
		//	os.Remove(file)
		}
	}
	service.Close()
}
