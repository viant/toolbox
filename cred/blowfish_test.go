package cred_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/cred"
	"testing"
)

func TestNewBlowfishCipher(t *testing.T) {
	cipher, err := cred.NewBlowfishCipher(cred.DefaultKey)
	if assert.Nil(t, err) {

		{
			var secret = "This is secret pass12312312321"
			encrypted := cipher.Encrypt([]byte(secret))
			decrypted := cipher.Decrypt(encrypted)
			assert.Equal(t, secret, string(decrypted))
		}

		{
			var secret = "abc"
			encrypted := cipher.Encrypt([]byte(secret))
			decrypted := cipher.Decrypt(encrypted)
			assert.Equal(t, secret, string(decrypted))
		}

		{
			var secret = "123!abc"
			encrypted := cipher.Encrypt([]byte(secret))
			decrypted := cipher.Decrypt(encrypted)
			assert.Equal(t, secret, string(decrypted))
		}

		{
			var secret = "test123@423 #!424"
			encrypted := cipher.Encrypt([]byte(secret))
			decrypted := cipher.Decrypt(encrypted)
			assert.Equal(t, secret, string(decrypted))
		}

		{
			var secret = "test123@423 #!424"
			encrypted := cipher.Encrypt([]byte(secret))
			decrypted := cipher.Decrypt(encrypted)
			assert.Equal(t, secret, string(decrypted))
		}

	}
}
