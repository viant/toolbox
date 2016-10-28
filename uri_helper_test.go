package toolbox_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestOpenURL(t *testing.T) {
	fileName, _, _ := toolbox.CallerInfo(2)
	{
		file, err := toolbox.OpenURL(toolbox.FileSchema+fileName, os.O_RDONLY, 0644)
		assert.Nil(t, err)
		defer file.Close()
	}
	{
		_, err := toolbox.OpenURL(toolbox.FileSchema+fileName+"bleh_bleh", os.O_RDONLY, 0644)
		assert.NotNil(t, err)
	}

	{
		_, err := toolbox.OpenURL("https://github.com/viant/toolbox", os.O_RDONLY, 0644)
		assert.NotNil(t, err, "only file protocol is supported")
	}

}

func TestOpenReaderFromURL(t *testing.T) {
	fileName, _, _ := toolbox.CallerInfo(2)
	{
		file, _, err := toolbox.OpenReaderFromURL(toolbox.FileSchema + fileName)
		assert.Nil(t, err)
		defer file.Close()
	}
	{
		_, _, err := toolbox.OpenReaderFromURL(toolbox.FileSchema + fileName + "blahbla")
		assert.NotNil(t, err)
	}

	{
		file, _, err := toolbox.OpenReaderFromURL("https://github.com/viant/toolbox")
		assert.Nil(t, err)
		defer file.Close()
	}

	{
		_, _, err := toolbox.OpenReaderFromURL("abc://github.com/viant/toolbox")
		assert.NotNil(t, err)
	}
}
