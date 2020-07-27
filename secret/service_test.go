package secret

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/storage"
	"testing"
	"time"
)

func setupData(baseDirectory string, data map[string]*cred.Config) error {
	storageService := storage.NewMemoryService()
	for k, v := range data {
		var buf = new(bytes.Buffer)
		if err := v.Write(buf); err != nil {
			return err
		}
		var URL = toolbox.URLPathJoin(baseDirectory, fmt.Sprintf("%v.json", k))
		if err := storageService.Upload(URL, bytes.NewReader(buf.Bytes())); err != nil {
			return err
		}
	}
	return nil
}

func TestNew(t *testing.T) {
	var baseDirectory = "mem://secret"
	err := setupData(baseDirectory, map[string]*cred.Config{
		"localhost": {Username: "user1", Password: "pass1"},
		"10.3.3.12": {Username: "xxx", Password: "yyy"},
	})
	assert.Nil(t, err)
	service := New(baseDirectory, false)

	config, err := service.GetCredentials("localhost")
	if assert.Nil(t, err) {
		assert.EqualValues(t, "user1", config.Username)
		assert.EqualValues(t, "pass1", config.Password)
	}
	_, err = service.GetCredentials("nonexisting")
	assert.NotNil(t, err)

}

func TestService_Expand(t *testing.T) {
	var baseDirectory = "mem://secret"
	err := setupData(baseDirectory, map[string]*cred.Config{
		"localhost":  {Username: "user1", Password: "pass1"},
		"10.3.3.12":  {Username: "xxx", Data: `{"Key1":"abc"}`},
		"github.com": {Username: "xxx", Password: "p"},
	})
	assert.Nil(t, err)
	service := New(baseDirectory, false)

	var original = ReadUserAndPassword
	defer func() {
		ReadUserAndPassword = original
	}()

	ReadUserAndPassword = func(timeout time.Duration) (string, string, error) {
		return "user1", "password2", nil
	}

	var useCases = []struct {
		Description string
		Interactive bool
		Matchable   string
		Input       string
		Credentials map[SecretKey]Secret
		Expect      string
		HasError    bool
	}{
		{
			Description: "Password replacement with secret short name",
			Input:       "**sudo**",
			Matchable:   "",
			Credentials: map[SecretKey]Secret{
				"**sudo**": "localhost",
			},
			Expect: "pass1",
		},
		{
			Description: "Password replacement with secret short name, explicit",
			Input:       "${sudo.password}",
			Matchable:   "",
			Credentials: map[SecretKey]Secret{
				"sudo": "localhost",
			},
			Expect: "pass1",
		},
		{
			Description: "Username replacement with URL based secret",
			Input:       "##sudo##",
			Matchable:   "",
			Credentials: map[SecretKey]Secret{
				"##sudo##": "mem://secret/localhost",
			},
			Expect: "user1",
		},
		{
			Description: "Username replacement with URL based secret and dynamic key",
			Input:       "##sudo##",
			Matchable:   "",
			Credentials: map[SecretKey]Secret{
				"sudo": "mem://secret/localhost",
			},
			Expect: "user1",
		},

		{
			Description: "Username replacement with URL based secret and dynamic key - explicit",
			Input:       "${sudo.username}",
			Matchable:   "",
			Credentials: map[SecretKey]Secret{
				"sudo": "mem://secret/localhost",
			},
			Expect: "user1",
		},

		{
			Description: "Non existing CredentialsFromLocation",
			Input:       "**key**",
			Matchable:   "",
			Credentials: map[SecretKey]Secret{
				"key": "abc",
			},
			HasError: true,
		},

		{
			Description: "Interactive credential",
			Interactive: true,
			Input:       "**key**",
			Matchable:   "",
			Credentials: map[SecretKey]Secret{
				"key": "test",
			},
			Expect: "password2",
		},
	}

	for _, useCase := range useCases {
		service.interactive = useCase.Interactive
		_, err := service.Expand(useCase.Input, useCase.Credentials)
		if useCase.HasError {
			assert.NotNil(t, err, useCase.Description)
			continue
		}
		if assert.Nil(t, err, useCase.Description) {
		}
	}
}

func TestService_Create(t *testing.T) {
	service := New("mem://secret", false)

	var original = ReadUserAndPassword
	defer func() {
		ReadUserAndPassword = original
	}()

	{ //read user and password
		ReadUserAndPassword = func(timeout time.Duration) (string, string, error) {
			return "user1", "password2", nil
		}
		location, err := service.Create("test", "")
		assert.Nil(t, err)

		assert.EqualValues(t, "mem://secret/test.json", location)
		service := New("", false)

		config, err := service.GetCredentials(location)
		if assert.Nil(t, err) {
			assert.EqualValues(t, "user1", config.Username)
			assert.EqualValues(t, "password2", config.Password)
		}

	}

	{ //Test interactive

		service.interactive = true
		ReadUserAndPassword = func(timeout time.Duration) (string, string, error) {
			return "user2", "password3", nil
		}
		config, err := service.GetOrCreate("xxxx")
		assert.Nil(t, err)
		if assert.Nil(t, err) {
			assert.EqualValues(t, "user2", config.Username)
			assert.EqualValues(t, "password3", config.Password)
		}

	}
	{ //read user and password with error
		ReadUserAndPassword = func(timeout time.Duration) (string, string, error) {
			return "", "", errors.New("test")
		}
		_, err := service.Create("test", "")
		assert.NotNil(t, err)
	}

}

func TestSecret_IsLocation(t *testing.T) {
	secret := Secret("mem://github.com/viant/endly/workflow/docker/build/secret/build.json")
	assert.True(t, secret.IsLocation())
}
