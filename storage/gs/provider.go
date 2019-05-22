package gs

import (
	"context"
	"encoding/json"
	"github.com/viant/toolbox/secret"
	"github.com/viant/toolbox/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"os"
)

const ProviderScheme = "gs"
const userAgent = "gcloud-golang-storage/20151204"
const DevstorageFullControlScope = "https://www.googleapis.com/auth/devstorage.full_control"
const googleStorageProjectKey = "GOOGLE_STORAGE_PROJECT"

func init() {
	storage.Registry().Registry[ProviderScheme] = serviceProvider
}

func serviceProvider(credentials string) (storage.Service, error) {
	var credentialOptions = make([]option.ClientOption, 0)
	var projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	if credentials == "" {
		credentialOptions = append([]option.ClientOption{},
			option.WithScopes(DevstorageFullControlScope),
			option.WithUserAgent(userAgent))
	} else {
		if json.Valid([]byte(credentials)) {
			credentialOptions = append(credentialOptions, option.WithCredentialsJSON([]byte(credentials)))
		} else {
			credentialOptions = append(credentialOptions, option.WithCredentialsFile(credentials))
		}
		secretService := secret.New("", false)
		config, err := secretService.GetCredentials(credentials)
		if err != nil {
			return nil, err
		}
		projectID = config.ProjectID
	}

	if customProjectID := os.Getenv(googleStorageProjectKey); customProjectID != "" {
		projectID = customProjectID
	}
	if projectID == "" {
		if credentials, err := google.FindDefaultCredentials(context.Background(), DevstorageFullControlScope); err == nil {
			projectID = credentials.ProjectID
		}
	}
	return NewService(projectID, credentialOptions...), nil
}
