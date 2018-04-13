package ssh_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/ssh"
	"path"
	"testing"
)

func Test_NewReplayService(t *testing.T) {

	parent := toolbox.CallerDirectory(3)
	commands, err := ssh.NewReplayCommands(path.Join(parent, "test/ls"))
	assert.Nil(t, err)
	err = commands.Load()
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, 3, len(commands.Commands))

	assert.Nil(t, err)
	defer commands.Store()
	service := ssh.NewReplayService("AWITAS-C02TF066H040:awitas1511796457759720702$", "darwin", commands, nil)
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
