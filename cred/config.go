package cred

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var sshKeyFileCandidates = []string{"/.ssh/id_rsa", "/.ssh/id_dsa"}
var DefaultKey = []byte{0x24, 0x66, 0xDD, 0x87, 0x8B, 0x96, 0x3C, 0x9D}
var PasswordCipher = GetDefaultPasswordCipher()

type Config struct {
	Username          string `json:",omitempty"`
	Password          string `json:",omitempty"`
	EncryptedPassword string `json:",omitempty"`
	PrivateKeyPath    string `json:",omitempty"`

	//amazon cloud credential
	Key    string `json:",omitempty"`
	Secret string `json:",omitempty"`
	Region string `json:",omitempty"`

	//google cloud credential
	ClientEmail  string `json:"client_email,omitempty"`
	TokenURL     string `json:"token_uri,omitempty"`
	PrivateKey   string `json:"private_key,omitempty"`
	PrivateKeyID string `json:"private_key_id,omitempty"`

	sshClientConfig *ssh.ClientConfig
	jwtClientConfig *jwt.Config
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

//IsKeyEncrypted checks if supplied key content is encrypyed by password
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

//SSHClientConfig returns a new instance of sshClientConfig
func (c *Config) SSHClientConfig() (*ssh.ClientConfig, error) {
	return c.ClientConfig()
}

//NewJWTConfig returns new JWT config for supplied scopes
func (c *Config) NewJWTConfig(scopes ...string) (*jwt.Config, error) {
	var result = &jwt.Config{
		Email:        c.ClientEmail,
		Subject:      c.ClientEmail,
		PrivateKey:   []byte(c.PrivateKey),
		PrivateKeyID: c.PrivateKeyID,
		Scopes:       scopes,
		TokenURL:     c.TokenURL,
	}
	if c.PrivateKeyPath != "" && c.PrivateKey == "" {
		privateKey, err := ioutil.ReadFile(c.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open provide key: %v, %v", c.PrivateKeyPath, err)
		}
		result.PrivateKey = privateKey
	}
	if result.TokenURL == "" {
		result.TokenURL = google.JWTTokenURL
	}
	return result, nil
}

//ClientConfig returns a new instance of sshClientConfig
func (c *Config) ClientConfig() (*ssh.ClientConfig, error) {
	if c.sshClientConfig != nil {
		return c.sshClientConfig, nil
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
	c.sshClientConfig = result
	return result, nil
}

//NewConfig create a new config for supplied file name
func NewConfig(filename string) (*Config, error) {
	var config = &Config{}
	err := config.Load(filename)
	if err != nil {
		return nil, err
	}
	config.applyDefaultIfNeeded()
	return config, nil
}

//GetDefaultPasswordCipher return a default password cipher
func GetDefaultPasswordCipher() Cipher {
	var result, err = NewBlowfishCipher(DefaultKey)
	if err != nil {
		return nil
	}
	return result
}
