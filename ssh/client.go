package ssh

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io"
	"path"
	"sync/atomic"
	"time"
	"github.com/viant/toolbox/cred"
	"net"
	"sync"
	"log"
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
	Forwarding []*Forwarding
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
		return fmt.Errorf("Failed to create session %v", err)
	}
	defer session.Close()

	writer, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("Failed to acquire stdin %v", err)
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
		return fmt.Errorf("Failed to start command%v %v",cmd, err)
	}
	createFileCommand := fmt.Sprintf("%v %d %s\n", createFileSequence, len(content), file)
	_, err = writer.Write([]byte(createFileCommand))
	if err != nil {
		return fmt.Errorf("Failed to write create file sequence: %v %v",content, err)
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
			return fmt.Errorf("Failed to write end transfer seq: %v", err)
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



func (c *Client) Close() error {
	if len(c.Forwarding) > 0 {
		for _, forwarding := range c.Forwarding {
			_ = forwarding.Close()
		}
	}

	return c.Client.Close()
}


//Forward forwards localAddress to remoteAddress on SSH connection
func (c *Client) Forward(localAddress, remoteAddress string) (error) {
	local, err := net.Listen("tcp", localAddress)
	if err != nil {
		return fmt.Errorf("Failed to listen on local: %v %v", localAddress, err)
	}
	var forwarding = NewForwarding(c.Client, remoteAddress, local)
	if len(c.Forwarding) == 0 {
		c.Forwarding = make([]*Forwarding ,0)
	}
	c.Forwarding = append(c.Forwarding, forwarding)
	go forwarding.Handle()
	return nil
}


//Forwarding represents a SSH forwarding link
type Forwarding struct {
	RemoteAddress string
	client *ssh.Client
	Local net.Listener
	Connections []net.Conn
	mutex *sync.Mutex
	closed int32
}





func (f *Forwarding) tunnelTraffic(localClient, remote  net.Conn) {
	defer localClient.Close()
	defer remote.Close()
	completionChannel := make(chan bool)
	go func() {
		_, err := io.Copy(localClient, remote)
		if err != nil {
			log.Printf("Failed to copy remote to local: %v", err)
		}
		completionChannel <- true
	}()


	go func() {
		_, err := io.Copy(remote, localClient)
		if err != nil {
			log.Printf("Failed to copy local to remote: %v", err)
		}
		completionChannel <- true
	}()
	<-completionChannel
}


//Handle wait for local client and forwards traffic
func (f *Forwarding) Handle() error {
	for {
		if atomic.LoadInt32(&f.closed) == 1 {
			return nil
		}
		localClient, err := f.Local.Accept()
		if err != nil {
			return err
		}
		remote, err := f.client.Dial("tcp", f.RemoteAddress)
		if err != nil {
			return fmt.Errorf("Failed to connect to remote: %v %v", f.RemoteAddress, err)
		}
		f.Connections = append(f.Connections, remote)
		f.Connections = append(f.Connections, localClient)
		go f.tunnelTraffic(localClient, remote)
	}
	return nil
}



//Close closes forwarding link
func (f *Forwarding) Close() error {
	atomic.StoreInt32(&f.closed, 1)
	_ =f.Local.Close()
	for _, remote := range f.Connections {
		_ = remote.Close()
	}
	return nil
}

//NewForwarding creates a new ssh forwarding link
func NewForwarding(client *ssh.Client, remoteAddress string, local net.Listener) *Forwarding {
		return &Forwarding{
			client:client,
			RemoteAddress:remoteAddress,
			Connections:make([]net.Conn, 0),
			Local:local,
			mutex:&sync.Mutex{},
		}
}

//NewClient create a new client, it takes host port and authentication config
func NewClient(host string, port int, authConfig *cred.Config) (*Client, error) {
	if authConfig == nil {
		authConfig = &cred.Config{}
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
