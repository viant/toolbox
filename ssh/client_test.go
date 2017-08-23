package ssh_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/ssh"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := ssh.NewClient("127.0.0.1", 22, nil)
	if err == nil {
		assert.NotNil(t, client)

		session, err := client.OpenMultiCommandSession(nil)
		assert.Nil(t, err)
		defer session.Close()

		assert.NotNil(t, session)

		var out string
		out, err = session.Run("ls /etc/hosts", 2000)
		assert.Equal(t, "/etc/hosts", out)

	} else {
		assert.Nil(t, client)
	}
}
