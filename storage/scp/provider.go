package scp

import (
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/ssh"
	"github.com/viant/toolbox/storage"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const ProviderScheme = "scp"

func init() {
	storage.NewStorageProvider().Registry[ProviderScheme] = serviceProvider
}

func serviceProvider(credentialFile string) (storage.Service, error) {
	if !strings.HasPrefix(credentialFile, "/") {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		credentialFile = path.Join(dir, credentialFile)
	}
	config := &ssh.AuthConfig{}
	err := toolbox.LoadConfigFromUrl("file://"+credentialFile, config)
	if err != nil {
		return nil, err
	}
	return NewService(config), nil
}
