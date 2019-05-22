package scp

import (
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/secret"
	"github.com/viant/toolbox/storage"
	"strings"
)

//ProviderScheme represents scp URL scheme for this provider
const ProviderScheme = "scp"

//SSHProviderScheme represents ssh URL scheme for this provider
const SSHProviderScheme = "ssh"

func init() {
	storage.Registry().Registry[ProviderScheme] = serviceProvider
	storage.Registry().Registry[SSHProviderScheme] = serviceProvider
}

func serviceProvider(credentials string) (storage.Service, error) {
	var config = &cred.Config{}
	if strings.TrimSpace(credentials) != "" {
		var err error
		secrets := secret.New("", false)
		config, err = secrets.GetCredentials(credentials)
		if err != nil {
			return nil, err
		}
	}
	return NewService(config), nil
}
