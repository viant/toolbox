package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"path"
)

const (
	createFileSequence = "C0644"
)

var endTransferSequence = []byte("\x00")

//Client represnt SSH client
type Client struct {
	*ssh.Client
}

//MultiCommandSession create a new MultiCommandSession
func (c *Client) OpenMultiCommandSession(config *SessionConfig) (*MultiCommandSession, error) {
	return newMultiCommandSession(c.Client, config)
}

func (c *Client) Run(comand string) error {
	session, err := c.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()
	return session.Run(comand)
}

//Upload uploads passed in content into remote destination
func (c *Client) Upload(destination string, content []byte) error {
	dir, file := path.Split(destination)
	if len(dir) > 0 {
		c.Run("mkdir -p " + dir)
	}
	session, err := c.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	writer, err := session.StdinPipe()
	if err != nil {
		return err
	}
	defer writer.Close()

	output, err := session.StdoutPipe()
	cmd := "scp -tr " + dir
	err = session.Start(cmd)
	if err != nil {
		return err
	}
	createFileCommand := fmt.Sprintf("%v %d %s\n", createFileSequence, len(content), file)
	_, err = writer.Write([]byte(createFileCommand))
	if err != nil {
		return err
	}
	_, err = writer.Write(content)
	if err != nil {
		return err
	}
	_, err = writer.Write(endTransferSequence)
	if err != nil {
		return err
	}

	buf := make([]byte, 128)
	_, err = output.Read(buf)
	return err
}

//Download download passed source file from remote host.
func (c *Client) Download(source string) ([]byte, error) {
	session, err := c.Client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.Output(fmt.Sprintf("cat %s", source))
}

//NewClient create a new client, it takes host port and authentication config
func NewClient(host string, port int, authConfig *AuthConfig) (*Client, error) {
	if authConfig == nil {
		authConfig = &AuthConfig{}
	}
	clientConfig, err := authConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	hostWithPort := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", hostWithPort, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial %v: %s", hostWithPort, err)
	}
	return &Client{
		Client: client,
	}, nil
}
