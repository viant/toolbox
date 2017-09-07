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

func TestClient_Upload(t *testing.T) {
	client, err := ssh.NewClient("127.0.0.1", 22, nil)
	if err == nil {
		assert.NotNil(t, client)
		err = client.Upload("/tmp/a/abcd.bin", []byte{0x1, 0x6, 0x10})
		assert.Nil(t, err)

		content, err := client.Download("/tmp/a/abcd.bin")
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x1, 0x6, 0x10}, (content))

	} else {
		assert.Nil(t, client)
	}
}



