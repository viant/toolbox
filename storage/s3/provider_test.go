package s3

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
)

const s3Url = "s3://path"
const s3Secret = "mock_s3.json"

//Troubleshooting prod issue with credentials not loading
func TestLoadingCredentialFile(t *testing.T) {
	// Creating absolute path
	fileName, _, _ := toolbox.CallerInfo(2)
	currentPath, _ := path.Split(fileName)
	credentialPath := path.Dir(path.Dir(currentPath)) + "/test/" + s3Secret
	service, err := storage.NewServiceForURL(s3Url, credentialPath)

	assert.Nil(t, err)
	assert.NotNil(t, service)

}
