package toolbox_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"net/url"
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

func Test_QueryValue(t *testing.T) {
	u, err := url.Parse("http://localhost/?k1=v1&k2=2&k3=false")
	assert.Nil(t, err)

	assert.Equal(t, "v1", toolbox.QueryValue(u, "k1", "default"))
	assert.Equal(t, "default", toolbox.QueryValue(u, "k10", "default"))

	assert.Equal(t, 2, toolbox.QueryIntValue(u, "k2", 3))
	assert.Equal(t, 3, toolbox.QueryIntValue(u, "k10", 3))

	assert.Equal(t, false, toolbox.QueryBoolValue(u, "k3", true))
	assert.Equal(t, true, toolbox.QueryBoolValue(u, "k10", true))

}
