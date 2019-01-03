package gs

import (
	"github.com/viant/toolbox/secret"
	"github.com/viant/toolbox/storage"
	"google.golang.org/api/option"
)

const ProviderScheme = "gs"

func init() {
	storage.NewStorageProvider().Registry[ProviderScheme] = serviceProvider
}

func serviceProvider(credentialsFile string) (storage.Service, error) {
	credentialOption := option.WithCredentialsFile(credentialsFile)
	secretService := secret.New("", false)
	config, err := secretService.GetCredentials(credentialsFile)
	if err != nil {
		return nil, err
	}
	return NewService(config.ProjectID, credentialOption), nil
}
