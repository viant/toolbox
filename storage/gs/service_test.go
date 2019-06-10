package gs

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/storage"
	"log"
	"os"
	"testing"
)

func TestService_List(t *testing.T) {


	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/awitas/.secret/viant-e2e.json")
	service, err := storage.NewServiceForURL("gs://doo/", "")
	if err != nil {
		log.Fatal(err)
	}

	reader, err := service.DownloadWithURL("gs://e2e-siteprofile-indexer/input/app_rules.csv")
	assert.Equal(t, reader, err)


	/*	credential := option.WithServiceAccountFile("<Secret file path>")
		service := gs.NewService(credential)
		assert.NotNil(t, service)
		objects, err := service.List("<GCS bucket>")
		assert.Nil(t, err)
		for _, o := range objects {
			fmt.Printf("%v\n", o.URL())
		}
		object, err := service.StorageObject("<GCS bucket>")
		assert.Nil(t, err)
		assert.NotNil(t, object)
		err = service.Delete(object)
		assert.Nil(t, err)*/

}
