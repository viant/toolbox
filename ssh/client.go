package ssh

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io"
	"path"
	"sync/atomic"
	"time"
)

const (
	createFileSequence = "C0644"
)

var bufferSize = 64 * 1024
var scpUploadSleep = 100 * time.Millisecond
var commandResponseDelaySleep = 200 * time.Millisecond

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

//listenForMessage this function read data from reader to filer textual output to result channel.
func listenForMessage(reader io.Reader, result chan string, done *int32) {
	for {
		if atomic.LoadInt32(done) == 1 {
			return
		}
		var buf = make([]byte, bufferSize)
		read, _ := reader.Read(buf)
		if read > 0 {

			data := buf[:read]
			var text = ""
			for _, b := range data {
				if b >= 32 {
					text += string(b)
				}
			}
			if text != "" {
				result <- text
			}
		}
	}
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

	var done int32
	defer func() {
		atomic.StoreInt32(&done, 1)
	}()
	output, err := session.StdoutPipe()
	var messages = make(chan string, 1)
	go listenForMessage(output, messages, &done)

	cmd := "scp -qtr " + dir
	err = session.Start(cmd)
	if err != nil {
		return err
	}
	createFileCommand := fmt.Sprintf("%v %d %s\n", createFileSequence, len(content), file)
	_, err = writer.Write([]byte(createFileCommand))
	if err != nil {
		return err
	}
	var message string
	select {
	case message = <-messages:
	case <-time.After(commandResponseDelaySleep):
	}
	if message != "" {
		return errors.New(message)
	}

	var interationCount = (len(content) / bufferSize) + 1
	//This is terrible hack, but  it looks like writer.Write at once or using io.Copy causes some data being lost in the final file,
	//so slowing down writes addresses this issue
	for i := 0; i < interationCount; i++ {
		maxLength := (i + 1) * bufferSize
		if maxLength >= len(content) {
			maxLength = len(content)
		}
		buffer := content[i*bufferSize: maxLength]
		_, err = writer.Write(buffer)
		if err != nil {
			if err.Error() == io.EOF.Error() {
				break
			}
			return fmt.Errorf("Failed to write content %v %v %v", err, len(content), i)
		}
		if i+2 > interationCount {
			time.Sleep(scpUploadSleep)
		}
	}

	if err == nil {
		_, err = writer.Write(endTransferSequence)
		if err != nil {
			return err
		}
	}
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
