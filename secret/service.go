package secret

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/data"
	"github.com/viant/toolbox/storage"
	"github.com/viant/toolbox/url"
	"os"
	"path"
	"strings"
	"sync"
)

//represents a secret service
type Service struct {
	interactive   bool
	baseDirectory string
	cache         map[string]*cred.Config
	lock          *sync.RWMutex
}

func (s *Service) CredentialsLocation(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("secretLocation was empty")
	}
	if path.Ext(secret) == "" {
		secret += ".json"
	}

	if strings.HasPrefix(secret, "~") {
		secret = strings.Replace(secret, "~", os.Getenv("HOME"), 1)
	}
	currentDirectory, _ := os.Getwd()
	for _, candidate := range []string{secret, toolbox.URLPathJoin(currentDirectory, secret)} {
		if toolbox.FileExists(candidate) {
			return candidate, nil
		}
	}
	if strings.Contains(secret, ":/") {
		return secret, nil
	}
	return toolbox.URLPathJoin(s.baseDirectory, secret), nil
}

//Credentials returns credential config for supplied location.
func (s *Service) CredentialsFromLocation(secret string) (*cred.Config, error) {
	secretLocation, err := s.CredentialsLocation(secret)
	if err != nil {
		return nil, err
	}
	s.lock.RLock()
	credConfig, has := s.cache[secretLocation]
	s.lock.RUnlock()
	if has {
		return credConfig, nil
	}
	resource := url.NewResource(secretLocation)
	configContent, err := resource.Download()
	if err != nil {
		return nil, fmt.Errorf("failed to open: '%v', due to: %v", secretLocation, err)
	}
	credConfig = &cred.Config{}
	if err = credConfig.LoadFromReader(bytes.NewReader(configContent), path.Ext(secretLocation)); err != nil {
		return nil, err
	}
	credConfig.Data = string(configContent)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache[secretLocation] = credConfig
	return credConfig, nil
}

//GetOrCreate gets or creates credential
func (s *Service) GetOrCreate(secret string) (*cred.Config, error) {
	if secret == "" {
		return nil, errors.New("secret was empty")
	}
	result, err := s.GetCredentials(secret)
	if err != nil && s.interactive && Secret(secret).IsLocation() {
		secretLocation, err := s.Create(secret, "")
		if err != nil {
			return nil, err
		}
		return s.GetCredentials(secretLocation)
	}
	return result, err
}

//Credentials returns credential config
func (s *Service) GetCredentials(secret string) (*cred.Config, error) {
	if !Secret(secret).IsLocation() {
		var result = &cred.Config{Data: string(secret)}
		//try to load credential
		err := result.LoadFromReader(strings.NewReader(string(secret)), "")
		return result, err
	}
	return s.CredentialsFromLocation(secret)
}

func (s *Service) expandDynamicSecret(input string, key SecretKey, secret Secret) (string, error) {
	if !strings.Contains(input, key.String()) {
		return input, nil
	}
	credConfig, err := s.GetOrCreate(string(secret))
	if err != nil {
		return "", err
	}
	createMap := data.NewMap()
	credInfo := map[string]interface{}{}
	_ = toolbox.DefaultConverter.AssignConverted(&credInfo, credInfo)
	credInfo["username"] = credConfig.Username
	credInfo["password"] = credConfig.Password
	createMap.Put(string(key), credInfo)

	passwordKey := fmt.Sprintf("**%v**", key)
	if count := strings.Count(input, passwordKey); count > 0 {
		secret := credConfig.Password
		if secret == "" {
			secret = credConfig.Data
		}
		input = strings.Replace(input, passwordKey, secret, count)
	}
	userKey := fmt.Sprintf("##%v##", key)
	if count := strings.Count(input, userKey); count > 0 {
		input = strings.Replace(input, userKey, credConfig.Username, count)
	}
	if index := strings.Index(input, "$"); index == -1 {
		return input, nil
	}

	return createMap.ExpandAsText(input), nil
}

func (s *Service) expandSecret(command string, key SecretKey, secret Secret) (string, error) {
	credConfig, err := s.GetOrCreate(string(secret))
	if err != nil {
		return "", err
	}
	command = strings.Replace(command, key.String(), key.Secret(credConfig), 1)
	return command, nil
}

//Expand expands input credential keys with actual CredentialsFromLocation
func (s *Service) Expand(input string, credentials map[SecretKey]Secret) (string, error) {
	if len(credentials) == 0 {
		return input, nil
	}
	var err error
	for k, v := range credentials {
		if strings.Contains(input, k.String()) {
			if k.IsDynamic() {
				input, err = s.expandDynamicSecret(input, k, v)
			} else {
				input, err = s.expandSecret(input, k, v)
			}
			if err != nil {
				return "", err
			}
		}
	}
	return input, nil
}

//Create creates a new credential config for supplied name
func (s *Service) Create(name, privateKeyPath string) (string, error) {
	if strings.HasPrefix(privateKeyPath, "~") {
		privateKeyPath = strings.Replace(privateKeyPath, "~", os.Getenv("HOME"), 1)
	}
	username, password, err := ReadUserAndPassword(ReadingCredentialTimeout)
	if err != nil {
		return "", err
	}
	config := &cred.Config{
		Username: username,
		Password: password,
	}
	if toolbox.FileExists(privateKeyPath) && !cred.IsKeyEncrypted(privateKeyPath) {
		config.PrivateKeyPath = privateKeyPath
	}
	var secretResource = url.NewResource(toolbox.URLPathJoin(s.baseDirectory, fmt.Sprintf("%v.json", name)))
	storageService, err := storage.NewServiceForURL(secretResource.URL, "")
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = config.Write(buf)
	if err != nil {
		return "", err
	}
	err = storageService.Upload(secretResource.URL, buf)
	return secretResource.URL, err
}

//NewSecretService creates a new secret service
func New(baseDirectory string, interactive bool) *Service {
	if baseDirectory == "" {
		baseDirectory = path.Join(os.Getenv("HOME"), ".secret")
	} else if strings.HasPrefix(baseDirectory, "~") {
		baseDirectory = strings.Replace(baseDirectory, "~", path.Join(os.Getenv("HOME"), ".secret"), 1)
	}
	return &Service{
		baseDirectory: baseDirectory,
		interactive:   interactive,
		cache:         make(map[string]*cred.Config),
		lock:          &sync.RWMutex{},
	}
}
