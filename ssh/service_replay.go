package ssh

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
)

//Service represents ssh service
type replayService struct {
	storage     map[string][]byte
	shellPrompt string
	system      string
	commands    *ReplayCommands
}

//Service returns a service wrapper
func (s *replayService) Client() *ssh.Client {
	return &ssh.Client{}
}

//OpenMultiCommandSession opens multi command session
func (s *replayService) OpenMultiCommandSession(config *SessionConfig) (MultiCommandSession, error) {
	return NewReplayMultiCommandSession(s.shellPrompt, s.system, s.commands), nil
}

//Run runs supplied command
func (s *replayService) Run(command string) error {
	if commands, ok := s.commands.Commands[command]; ok {
		if commands.Error != "" {
			return errors.New(commands.Error)
		}
	}
	s.commands.Next(command)
	return nil
}

//Upload uploads provided content to specified destination
func (s *replayService) Upload(destination string, mode os.FileMode, content []byte) error {
	s.storage[destination] = content
	return nil
}

//Download downloads content from specified source.
func (s *replayService) Download(source string) ([]byte, error) {
	if _, has := s.storage[source]; !has {
		return nil, fmt.Errorf("no such file or directory")
	}
	return s.storage[source], nil
}

//OpenTunnel opens a tunnel between local to remote for network traffic.
func (s *replayService) OpenTunnel(localAddress, remoteAddress string) error {
	return nil
}

func (s *replayService) NewSession() (*ssh.Session, error) {
	return &ssh.Session{}, nil
}

func (s *replayService) Reconnect() error {
	return errors.New("unsupported")
}

func (s *replayService) Close() error {
	return nil
}

func NewReplayService(shellPrompt, system string, commands *ReplayCommands, storage map[string][]byte) Service {
	if len(storage) == 0 {
		storage = make(map[string][]byte)
	}
	return &replayService{
		storage:     storage,
		shellPrompt: shellPrompt,
		system:      system,
		commands:    commands,
	}
}
