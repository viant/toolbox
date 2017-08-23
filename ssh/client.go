package ssh

import (
	"fmt"
	"github.com/viant/toolbox"
	"golang.org/x/crypto/ssh"
	"net/http/httputil"
)

//Client represnt SSH client
type Client struct {
	*ssh.Client
	Pool httputil.BufferPool
}

//MultiCommandSession create a new MultiCommandSession
func (c *Client) OpenMultiCommandSession(config *SessionConfig) (*MultiCommandSession, error) {
	return newMultiCommandSession(c.Client, config)
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
		Pool:   toolbox.NewBytesBufferPool(10, 64*1024),
	}, nil
}
