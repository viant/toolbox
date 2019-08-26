package ssh

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/storage"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type (
	//Service represents ssh service
	Service interface {
		//Service returns a service wrapper
		Client() *ssh.Client

		//OpenMultiCommandSession opens multi command session
		OpenMultiCommandSession(config *SessionConfig) (MultiCommandSession, error)

		//Run runs supplied command
		Run(command string) error

		//Upload uploads provided content to specified destination
		//Deprecated: please consider using https://github.com/viant/afs/tree/master/scp
		Upload(destination string, mode os.FileMode, content []byte) error

		//Download downloads content from specified source.
		//Deprecated: please consider using https://github.com/viant/afs/tree/master/scp
		Download(source string) ([]byte, error)

		//OpenTunnel opens a tunnel between local to remote for network traffic.
		OpenTunnel(localAddress, remoteAddress string) error

		NewSession() (*ssh.Session, error)

		Close() error
	}
)

//service represnt SSH service
type service struct {
	host           string
	client         *ssh.Client
	forwarding     []*Tunnel
	replayCommands *ReplayCommands
	recordSession  bool
	config         *ssh.ClientConfig
}

//Service returns undelying ssh Service
func (c *service) Client() *ssh.Client {
	return c.client
}

//Service returns undelying ssh Service
func (c *service) NewSession() (*ssh.Session, error) {
	return c.client.NewSession()
}

//MultiCommandSession create a new MultiCommandSession
func (c *service) OpenMultiCommandSession(config *SessionConfig) (MultiCommandSession, error) {
	return newMultiCommandSession(c, config, c.replayCommands, c.recordSession)
}

func (c *service) Run(command string) error {
	session, err := c.client.NewSession()
	if err != nil {
		panic("failed to create session: " + err.Error())
	}
	defer session.Close()
	return session.Run(command)
}

func (c *service) transferData(payload []byte, createFileCmd string, writer io.Writer, errors chan error, waitGroup *sync.WaitGroup) {
	const endSequence = "\x00"
	defer waitGroup.Done()
	_, err := fmt.Fprint(writer, createFileCmd)
	if err != nil {
		errors <- err
		return
	}
	_, err = io.Copy(writer, bytes.NewReader(payload))
	if err != nil {
		errors <- err
		return
	}
	if _, err = fmt.Fprint(writer, endSequence); err != nil {
		errors <- err
		return
	}
}

type Errors chan error

func (e Errors) GetError() error {
	select {
	case err := <-e:
		return err
	case <-time.After(time.Millisecond):
	}
	return nil
}

const operationSuccessful = 0

func checkOutput(reader io.Reader, errorChannel Errors) {
	writer := new(bytes.Buffer)
	io.Copy(writer, reader)
	if writer.Len() > 1 {
		data := writer.Bytes()
		if data[1] == operationSuccessful {
			return
		} else if len(data) > 2 {
			errorChannel <- errors.New(string(data[2:]))
		}
	}
}

//Upload uploads passed in content into remote destination
func (c *service) Upload(destination string, mode os.FileMode, content []byte) (err error) {
	err = c.upload(destination, mode, content)

	if err != nil {
		if strings.Contains(err.Error(), "No such file or directory") {
			dir, _ := path.Split(destination)
			c.Run("mkdir -p " + dir)
			return c.upload(destination, mode, content)
		} else if strings.Contains(err.Error(), "handshake") || strings.Contains(err.Error(), "connection") {

			time.Sleep(500 * time.Millisecond)
			fmt.Printf("got error %v\n", err)
			c.Reconnect()
			return c.upload(destination, mode, content)
		}
	}
	return err
}

func (c *service) getSession() (*ssh.Session, error) {
	return c.client.NewSession()
}

//Upload uploads passed in content into remote destination
func (c *service) upload(destination string, mode os.FileMode, content []byte) (err error) {
	dir, file := path.Split(destination)
	if mode == 0 {
		mode = 0644
	}
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	if strings.HasPrefix(file, "/") {
		file = string(file[1:])
	}
	session, err := c.getSession()
	if err != nil {
		return err
	}

	writer, err := session.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "failed to acquire stdin")
	}
	defer writer.Close()

	var transferError Errors = make(chan error, 1)
	defer close(transferError)
	var sessionError Errors = make(chan error, 1)
	defer close(sessionError)
	output, err := session.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to acquire stdout")
	}
	go checkOutput(output, sessionError)

	if mode >= 01000 {
		mode = storage.DefaultFileMode
	}
	fileMode := string(fmt.Sprintf("C%04o", mode)[:5])
	createFileCmd := fmt.Sprintf("%v %d %s\n", fileMode, len(content), file)
	go c.transferData(content, createFileCmd, writer, transferError, waitGroup)
	scpCommand := "scp -qtr " + dir
	err = session.Start(scpCommand)
	if err != nil {
		return err
	}
	waitGroup.Wait()
	writerErr := writer.Close()
	if err := sessionError.GetError(); err != nil {
		return err
	}
	if err := transferError.GetError(); err != nil {
		return err
	}
	if err = session.Wait(); err != nil {
		if err := sessionError.GetError(); err != nil {
			return err
		}
		return err
	}
	return writerErr
}

//Download download passed source file from remote host.
func (c *service) Download(source string) ([]byte, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.Output(fmt.Sprintf("cat %s", source))
}

//Host returns client host
func (c *service) Host() string {
	return c.host
}

//Close closes service
func (c *service) Close() error {
	if len(c.forwarding) > 0 {
		for _, forwarding := range c.forwarding {
			_ = forwarding.Close()
		}
	}
	return c.client.Close()
}

//Reconnect client
func (c *service) Reconnect() error {
	return c.connect()
}

//OpenTunnel tunnels data between localAddress and remoteAddress on ssh connection
func (c *service) OpenTunnel(localAddress, remoteAddress string) error {
	local, err := net.Listen("tcp", localAddress)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to listen on local: %v %v", localAddress))
	}
	var forwarding = NewForwarding(c.client, remoteAddress, local)
	if len(c.forwarding) == 0 {
		c.forwarding = make([]*Tunnel, 0)
	}
	c.forwarding = append(c.forwarding, forwarding)
	go forwarding.Handle()
	return nil
}

func (c *service) connect() (err error) {
	if c.client, err = ssh.Dial("tcp", c.host, c.config); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to dial %v: %s", c.host))
	}
	return nil
}

//NewService create a new ssh service, it takes host port and authentication config
func NewService(host string, port int, authConfig *cred.Config) (Service, error) {
	if authConfig == nil {
		authConfig = &cred.Config{}
	}
	clientConfig, err := authConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	var result = &service{
		host:   fmt.Sprintf("%s:%d", host, port),
		config: clientConfig,
	}
	return result, result.connect()
}
