package cred_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/cred"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	var tempDir = os.TempDir()
	var testFile = path.Join(tempDir, "credTest1.json")
	_ = os.Remove(testFile)
	var data = "{\"Username\":\"adrian\", \"Password\":\"abc\"}"
	err := ioutil.WriteFile(testFile, []byte(data), 0644)
	assert.Nil(t, err)
	{
		config, err := cred.NewConfig(testFile)
		assert.Nil(t, err)
		assert.Equal(t, "abc", config.Password)
		assert.Equal(t, "adrian", config.Username)
		assert.Equal(t, "AAAAAAAAAAAXUPcVbxwWlQ==", config.EncryptedPassword)
		_ = os.Remove(testFile)
		config.Save(testFile)
	}

	{
		config, err := cred.NewConfig(testFile)
		assert.Nil(t, err)
		assert.Equal(t, "abc", config.Password)
		assert.Equal(t, "adrian", config.Username)
		assert.Equal(t, "AAAAAAAAAAAXUPcVbxwWlQ==", config.EncryptedPassword)
	}

	{
		configJSON := `{"Username":"adrian","EncryptedPassword":"AAAAAAAAAAAXUPcVbxwWlQ=="}`

		config := cred.Config{}
		err = config.LoadFromReader(strings.NewReader(configJSON), ".json")
		assert.Nil(t, err)

		assert.EqualValues(t, "abc", config.Password)
	}

	_ = os.Remove(testFile)

}
