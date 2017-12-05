package url_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/url"
	"os"
	"path"
	"testing"
)

func TestNewResource(t *testing.T) {

	{
		var resource = url.NewResource("https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt")
		assert.EqualValues(t, resource.ParsedURL.String(), "https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt")
		data, err := resource.Download()
		assert.Nil(t, err)
		assert.NotNil(t, data)
	}

	var resource = url.NewResource("https://raw.githubusercontent.com/viant//toolbox//master/LICENSE.txt")
	assert.Equal(t, "https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt", resource.URL)
}

func TestNew_CredentialURL(t *testing.T) {

	{
		var resource = url.NewResource("https://raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt?check=1&p=2")
		var URL = resource.CredentialURL("smith", "123")
		assert.EqualValues(t, "https://smith:123@raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt?check=1&p=2", URL)
	}

	{
		var resource = url.NewResource("https://raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt")
		var URL = resource.CredentialURL("smith", "")
		assert.EqualValues(t, "https://smith@raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt", URL)
	}

}

func TestNew_DirectoryPath(t *testing.T) {
	{
		var resource = url.NewResource("https://raw.githubusercontent.com:80/viant/toolbox/master/LICENSE.txt")
		assert.EqualValues(t, "/viant/toolbox/master", resource.DirectoryPath())
	}
	{
		var resource = url.NewResource("https://raw.githubusercontent.com:80/viant/toolbox/master/avc")
		assert.EqualValues(t, "/viant/toolbox/master/avc", resource.DirectoryPath())
	}
	{
		var resource = url.NewResource("hter")
		assert.EqualValues(t, "", resource.DirectoryPath())
	}
}

func TestResource_YamlDecode(t *testing.T) {
	var filename = path.Join(os.Getenv("TMPDIR"), "resource.yaml")
	toolbox.RemoveFileIfExist(filename)
	defer toolbox.RemoveFileIfExist(filename)
	var aMap = map[string]interface{}{
		"a": 1,
		"b": "123",
		"c": []int{1, 3, 6},
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	fmt.Printf("%v\n", filename)
	if assert.Nil(t, err) {
		err = toolbox.NewYamlEncoderFactory().Create(file).Encode(aMap)
		assert.Nil(t, err)
	}

	var resource = url.NewResource(filename)
	assert.EqualValues(t, resource.ParsedURL.String(), toolbox.FileSchema+filename)

	var resourceData = make(map[string]interface{})
	err = resource.YAMLDecode(&resourceData)
	assert.Nil(t, err)

	assert.EqualValues(t, resourceData["a"], 1)
	assert.EqualValues(t, resourceData["b"], "123")

}

func TestResource_JsonDecode(t *testing.T) {
	var filename = path.Join(os.Getenv("TMPDIR"), "resource.json")
	toolbox.RemoveFileIfExist(filename)
	defer toolbox.RemoveFileIfExist(filename)
	var aMap = map[string]interface{}{
		"a": 1,
		"b": "123",
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	fmt.Printf("%v\n", filename)
	if assert.Nil(t, err) {
		err = toolbox.NewJSONEncoderFactory().Create(file).Encode(aMap)
		assert.Nil(t, err)
	}

	var resource = url.NewResource(filename)
	assert.EqualValues(t, resource.ParsedURL.String(), toolbox.FileSchema+filename)

	var resourceData = make(map[string]interface{})
	err = resource.JSONDecode(&resourceData)
	assert.Nil(t, err)

	assert.EqualValues(t, resourceData["a"], 1)
	assert.EqualValues(t, resourceData["b"], "123")

}

func TestResource_LoadCredential(t *testing.T) {

	{
		var filename = path.Join(os.Getenv("TMPDIR"), "resource_secret.json")
		toolbox.RemoveFileIfExist(filename)
		defer toolbox.RemoveFileIfExist(filename)
		var aMap = map[string]interface{}{
			"username": "uzytkownik",
			"password": "haslo",
		}
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
		fmt.Printf("%v\n", filename)
		if assert.Nil(t, err) {
			err = toolbox.NewJSONEncoderFactory().Create(file).Encode(aMap)
			assert.Nil(t, err)
		}
		var resource = url.NewResource("https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt", filename)

		username, password, err := resource.LoadCredential(false)
		assert.Nil(t, err)
		assert.Equal(t, username, "uzytkownik")
		assert.Equal(t, password, "haslo")
	}

	{ //error case

		var resource = url.NewResource("https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt")
		_, _, err := resource.LoadCredential(true)
		assert.NotNil(t, err)

	}

	{ //error case

		var resource = url.NewResource("https://raw.githubusercontent.com/viant/toolbox/master/LICENSE.txt", "bogus 343")
		_, _, err := resource.LoadCredential(true)
		assert.NotNil(t, err)

	}

}
