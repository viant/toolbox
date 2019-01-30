package gs

import (
	"github.com/viant/toolbox/secret"
	"github.com/viant/toolbox/storage"
	"google.golang.org/api/option"
	"os"
)

const ProviderScheme = "gs"
const userAgent = "gcloud-golang-storage/20151204"
const DevstorageFullControlScope = "https://www.googleapis.com/auth/devstorage.full_control"
const googleStorageProjectKey = "GOOGLE_STORAGE_PROJECT"

func init() {
	storage.NewStorageProvider().Registry[ProviderScheme] = serviceProvider
}

func serviceProvider(credentialsFile string) (storage.Service, error) {
	var credentialOptions = make([]option.ClientOption, 0)
	var projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	if credentialsFile == "" {
		credentialOptions = append([]option.ClientOption{},
			option.WithScopes(DevstorageFullControlScope),
			option.WithUserAgent(userAgent))
	} else {
		credentialOption := option.WithCredentialsFile(credentialsFile)
		credentialOptions = append(credentialOptions, credentialOption)
		secretService := secret.New("", false)
		config, err := secretService.GetCredentials(credentialsFile)
		if err != nil {
			return nil, err
		}
		projectID = config.ProjectID
	}

	if customProjectID := os.Getenv(googleStorageProjectKey); customProjectID != "" {
		projectID = customProjectID
	}
	return NewService(projectID, credentialOptions...), nil
}
