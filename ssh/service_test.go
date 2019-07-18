package ssh_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/ssh"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestNewClient(t *testing.T) {
	commands, err := ssh.NewReplayCommands("/tmp/ls")
	if err != nil {
		return
	}

	defer commands.Store()
	service, err := ssh.NewService("127.0.0.1", 22, nil)
	if err != nil {
		return
	}

	assert.NotNil(t, service)
	commands.Enable(service)

	if err == nil {
		assert.NotNil(t, service)
		session, err := service.OpenMultiCommandSession(nil)
		assert.Nil(t, err)
		defer session.Close()

		assert.NotNil(t, session)
		var out string
		out, err = session.Run("ls /etc/hosts", nil, 2000)
		assert.Equal(t, "/etc/hosts", out)

	} else {
		assert.Nil(t, service)
	}

}

func TestClient_Upload(t *testing.T) {
	service, err := ssh.NewService("127.0.0.1", 22, nil)
	if err != nil {
		return
	}
	if err == nil {
		assert.NotNil(t, service)
		err = service.Upload("/tmp/a/abcd.bin", 0644, []byte{0x1, 0x6, 0x10})
		assert.Nil(t, err)

		content, err := service.Download("/tmp/a/abcd.bin")
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x1, 0x6, 0x10}, (content))

	} else {
		assert.Nil(t, service)
	}
}

func TestClient_Key(t *testing.T) {
	auth, err := cred.NewConfig("/Users/awitas/.secret/scp1.json")
	if err != nil {
		return
	}
	assert.Nil(t, err)
	service, err := ssh.NewService("127.0.0.1", 22, auth)
	if err != nil {
		return
	}
	assert.NotNil(t, service)
}

func TestClient_UploadLargeFile(t *testing.T) {

	service, err := ssh.NewService("127.0.0.1", 22, nil)
	if err != nil {
		return
	}

	tempdir := os.TempDir()
	filename := path.Join(tempdir, "kkk/.bin/largefile.bin")
	toolbox.RemoveFileIfExist(filename)
	//defer toolbox.RemoveFileIfExist(filename)
	//file, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE, 0644)
	//assert.Nil(t, err)

	var payload = make([]byte, 1024*1024*16)
	for i := 0; i < len(payload); i += 32 {
		payload[i] = byte(i % 256)
	}
	//file.Write(payload)
	//file.Close()

	err = service.Upload(filename, 0644, payload)
	fmt.Printf("%v\n", err)
	assert.Nil(t, err)

	data, err := ioutil.ReadFile(filename)
	assert.Nil(t, err)
	if assert.Equal(t, len(payload), len(data)) {
		assert.EqualValues(t, data, payload)
	}

}
