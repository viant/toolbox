package cred_test

import (
	"testing"
	"github.com/viant/toolbox/cred"
	"github.com/stretchr/testify/assert"
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
	}
}


