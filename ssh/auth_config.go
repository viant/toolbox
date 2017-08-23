package ssh

import (
	"github.com/viant/toolbox"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path"
)

var sshKeyFileCandidates = []string{"/.ssh/id_rsa", "/.ssh/id_dsa"}

//AuthConfig represent SSH authentication config
type AuthConfig struct {
	Username       string
	Password       string
	PrivateKeyPath string
	clientConfig   *ssh.ClientConfig
}

func (c *AuthConfig) applyDefaultIfNeeded() {
	if c.Username == "" {
		c.Username = os.Getenv("USER")
	}
	if c.PrivateKeyPath == "" && c.Password == "" {
		homeDirectory := os.Getenv("HOME")
		if homeDirectory != "" {
			for _, candidate := range sshKeyFileCandidates {
				file := path.Join(homeDirectory, candidate)
				if toolbox.FileExists(file) {
					c.PrivateKeyPath = file
					break
				}
			}
		}
	}
}

//ClientConfig created a new instace of sshClientConfig
func (c *AuthConfig) ClientConfig() (*ssh.ClientConfig, error) {
	if c.clientConfig != nil {
		return c.clientConfig, nil
	}
	c.applyDefaultIfNeeded()
	result := &ssh.ClientConfig{
		User:            c.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            make([]ssh.AuthMethod, 0),
	}
	if c.Password != "" {
		result.Auth = append(result.Auth, ssh.Password(c.Password))
	} else if c.PrivateKeyPath != "" {
		privateKeyBytes, err := ioutil.ReadFile(c.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		key, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			return nil, err
		}
		result.Auth = append(result.Auth, ssh.PublicKeys(key))

	}
	c.clientConfig = result
	return result, nil
}
