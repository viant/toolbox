package gs

import (
	"github.com/viant/toolbox/storage"
	"google.golang.org/api/option"
)

const ProviderScheme = "gc"

func init() {
	storage.NewStorageProvider().Registry[ProviderScheme] = serviceProvider
}

func serviceProvider(credentialFile string) (storage.Service, error) {
	credentialOption := option.WithServiceAccountFile(credentialFile)
	return NewService(credentialOption), nil
}
