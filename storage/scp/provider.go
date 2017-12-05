package scp

import (
	"github.com/viant/toolbox/cred"
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
	var config = &cred.Config{}
	if credentialFile != "" {

		if !strings.HasPrefix(credentialFile, "/") {
			dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			credentialFile = path.Join(dir, credentialFile)
		}
		var err error
		config, err = cred.NewConfig(credentialFile)
		if err != nil {
			return nil, err
		}
	}
	return NewService(config), nil
}
