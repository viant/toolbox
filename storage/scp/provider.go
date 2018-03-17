package scp

import (
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/storage"
	"github.com/viant/toolbox/secret"
	"strings"
)

//ProviderScheme represents scp URL scheme for this provider
const ProviderScheme = "scp"
//SSHProviderScheme represents ssh URL scheme for this provider
const SSHProviderScheme = "ssh"

func init() {
	storage.NewStorageProvider().Registry[ProviderScheme] = serviceProvider
	storage.NewStorageProvider().Registry[SSHProviderScheme] = serviceProvider
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
