package s3

import (
	"github.com/viant/toolbox/cred"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/viant/toolbox/storage"
	"github.com/viant/toolbox/url"
)

const ProviderScheme = "s3"

func init() {
	SetDefaultProvider()
}

func serviceProvider(credentialFile string) (storage.Service, error) {
	s3config := &cred.Config{}
	if credentialFile != "" {
		if !strings.HasPrefix(credentialFile, "/") {
			dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			credentialFile = path.Join(dir, credentialFile)
		}
		resource := url.NewResource(credentialFile)
		err := resource.Decode(s3config)
		if err != nil {
			return nil, err
		}
	}
	return NewService(s3config), nil
}

//SetProvider set s3 provider with dynamic credentials
func SetDefaultProvider() {
	storage.Registry().Registry[ProviderScheme] = serviceProvider
}

//SetProvider set s3 provider with supplied config
func SetProvider(config *cred.Config) {
	storage.Registry().Registry[ProviderScheme] = func(string) (storage.Service, error) {
		return NewService(config), nil
	}
}
