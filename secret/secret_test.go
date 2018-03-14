package secret_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/secret"
	"testing"
)

func TestSecret_IsLocation(t *testing.T) {

	{
		sec := secret.Secret("htttp://abc.com/secret.json")
		assert.True(t, sec.IsLocation())
	}
	{
		sec := secret.Secret("{}")
		assert.False(t, sec.IsLocation())
	}

}

func TestSecretKey_IsDynamic(t *testing.T) {

	{
		key := secret.SecretKey("**key1**")
		assert.False(t, key.IsDynamic())
	}
	{
		key := secret.SecretKey("##key1##")
		assert.False(t, key.IsDynamic())
	}
	{
		key := secret.SecretKey("key1")
		assert.True(t, key.IsDynamic())
	}

}

func TestSecretKey_Secret(t *testing.T) {

	{ //test user and password secret
		credConfig := &cred.Config{Username: "abc", Password: "pass"}
		{
			key := secret.SecretKey("**key1**")
			assert.EqualValues(t, "pass", key.Secret(credConfig))
		}
		{
			key := secret.SecretKey("##key1##")
			assert.EqualValues(t, "abc", key.Secret(credConfig))
		}
	}

	{ //test json secret
		credConfig := &cred.Config{Username: "abc", Data: "{}"}
		{
			key := secret.SecretKey("**key1**")
			assert.EqualValues(t, "{}", key.Secret(credConfig))
		}
	}

}
