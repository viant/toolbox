package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
)

func TestCreateDirIfNotExist(t *testing.T) {
	dir := "/tmp/abc"
	toolbox.RemoveFileIfExist(dir)
	toolbox.CreateDirIfNotExist(dir)
	assert.True(t, toolbox.FileExists(dir))
	toolbox.RemoveFileIfExist(dir)
	assert.False(t, toolbox.FileExists(dir))
}
