package secret

import (
	"strings"
	"github.com/viant/toolbox/cred"
	"sync"
	"fmt"
	"bytes"
	"path"
	"errors"
	"github.com/viant/toolbox/url"
	"github.com/viant/toolbox"
	"os"
	"github.com/viant/toolbox/storage"
)

//represents a secret service
type Service struct {
	interactive   bool
	baseDirectory string
	cache         map[string]*cred.Config
	lock          *sync.RWMutex
}

//Credentials returns credential config for supplied location.
func (s *Service) credentials(secret string) (*cred.Config, error) {
	if secret == "" {
		return nil, errors.New("secretLocation was empty")
	}

	secretLocation := secret
	if ! (strings.Contains(secret, "://") || strings.HasPrefix(secret, "/")) {
		secretLocation = toolbox.URLPathJoin(s.baseDirectory, secret)
	}
	if path.Ext(secretLocation) == "" {
		secretLocation += ".json"
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
		return nil, fmt.Errorf("failed to open: %v", secretLocation)
	}
	credConfig = &cred.Config{}
	if err = credConfig.LoadFromReader(bytes.NewReader(configContent), path.Ext(secretLocation)); err != nil {
		return nil, err
	}
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
	if s.interactive &&  err != nil && Secret(secret).IsLocation() {
		secretLocation, err := s.Create(secret, "")
		if err != nil {
			return nil, err
		}
		return s.GetCredentials(secretLocation)
	}
	return result, err
}


//Credential returns credential config
func (s *Service) GetCredentials(secret string) (*cred.Config, error) {
	if ! Secret(secret).IsLocation() {
		var result = &cred.Config{Data: string(secret)}
		//try to load credential
		result.LoadFromReader(strings.NewReader(string(secret)), "")
		return result, nil
	}
	secretLocation := string(secret)
	return s.credentials(secretLocation);
}

func (s *Service) expandDynamicSecret(input string, key SecretKey, secret Secret) (string, error) {
	if ! strings.Contains(input, key.String()) {
		return input, nil
	}
	var err error
	for _, candidate := range key.Keys() {
		if input, err = s.expandSecret(input, candidate, secret); err != nil {
			return input, err
		}
	}
	return input, nil
}

func (s *Service) expandSecret(command string, key SecretKey, secret Secret) (string, error) {
	credConfig, err := s.GetOrCreate(string(secret))
	if err != nil {
		return "", err
	}
	command = strings.Replace(command, key.String(), key.Secret(credConfig), 1)
	return command, nil
}

//Expand expands input credential keys with actual credentials
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
	fmt.Printf("Credentials %v\n", name)
	username, password, err := ReadUserAndPassword(ReadingCredentialTimeout)
	if err != nil {
		return "", err
	}
	fmt.Println("")
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
