package cred

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"encoding/pem"
	"fmt"
)

var sshKeyFileCandidates = []string{"/.ssh/id_rsa", "/.ssh/id_dsa"}
var DefaultKey = []byte{0x24, 0x66, 0xDD, 0x87, 0x8B, 0x96, 0x3C, 0x9D}
var PasswordCipher = GetDefaultPasswordCipher()

type Config struct {
	Username          string
	Password          string
	EncryptedPassword string
	PrivateKeyPath    string
	Region            string
	Key               string
	Secret            string
	clientConfig      *ssh.ClientConfig
}

func (c *Config) Load(filename string) error {
	reader, err := os.Open(filename)
	if err != nil {
		return err
	}
	ext := path.Ext(filename)
	if strings.Contains(ext, "yaml") || strings.Contains(ext, "yml") {
		var data, err = ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(data, c)
		if err != nil {
			return err
		}
	} else {
		err = json.NewDecoder(reader).Decode(c)
		if err != nil {
			return nil
		}
	}

	if c.EncryptedPassword != "" {
		decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(c.EncryptedPassword))
		data, err := ioutil.ReadAll(decoder)
		if err != nil {
			return err
		}
		c.Password = string(PasswordCipher.Decrypt(data))
	} else if c.Password != "" {
		c.encryptPassword(c.Password)
	}
	return nil
}

func (c *Config) Save(filename string) error {
	var password = c.Password
	defer func() { c.Password = password }()
	if password != "" {
		c.encryptPassword(password)
		c.Password = ""
	}

	_ = os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(c)
}
func (c *Config) encryptPassword(password string) {

	encrypted := PasswordCipher.Encrypt([]byte(password))
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	defer encoder.Close()
	encoder.Write(encrypted)
	encoder.Close()
	c.EncryptedPassword = string(buf.Bytes())
}

func (c *Config) applyDefaultIfNeeded() {
	if c.Username == "" {
		c.Username = os.Getenv("USER")
	}
	if c.PrivateKeyPath == "" && c.Password == "" {
		homeDirectory := os.Getenv("HOME")
		if homeDirectory != "" {
			for _, candidate := range sshKeyFileCandidates {
				filename := path.Join(homeDirectory, candidate)
				file, err := os.Open(filename)
				if err == nil {
					file.Close()
					c.PrivateKeyPath = filename
					break
				}
			}
		}
	}
}

func IsKeyEncrypted(keyPath string) bool {
	privateKeyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return false
	}
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return false
	}
	return strings.Contains(block.Headers["Proc-Type"], "ENCRYPTED")
}



//ClientConfig returns a new instance of sshClientConfig
func (c *Config) ClientConfig() (*ssh.ClientConfig, error) {
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
	}
	if c.PrivateKeyPath != "" {

		if IsKeyEncrypted(c.PrivateKeyPath) {
			return nil, fmt.Errorf("key: %v, has been encrypeed with password", c.PrivateKeyPath)
		}

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

func NewConfig(filename string) (*Config, error) {
	var config = &Config{}
	err := config.Load(filename)
	if err != nil {
		return nil, err
	}
	config.applyDefaultIfNeeded()
	return config, nil
}

func GetDefaultPasswordCipher() Cipher {
	var result, err = NewBlowfishCipher(DefaultKey)
	if err != nil {
		return nil
	}
	return result
}
