package gs_test

import (
	"github.com/stretchr/testify/assert"
	"github.vianttech.com/core-adservers/cloud/storage/gs"
	"google.golang.org/api/option"
	"testing"
)

func TestService_List(t *testing.T) {

	credential := option.WithServiceAccountFile("/Users/awitas/Adelphic-af77adaff66a.json")
	service := gs.NewService(credential)
	assert.Nil(t, service)

	//objects, err := service.List("gs://s3adlogs/ad.log.go")
	//assert.Nil(t, err)
	//assert.Equal(t, 1, len(objects));

	//_, err := service.Download(objects[0])
	//assert.Nil(t, err)

	//content, err := ioutil.ReadAll(reader)
	//assert.Nil(t, err)
	//fmt.Printf("%v\n", string(content))
	//assert.True(t, len(content) > 0)
	//err = service.Upload("gs://s3adlogs/ad1.log?expiry=10", bytes.NewReader([]byte("abc")))
	//assert.Nil(t, err)
	//
	//object, err := service.Object("gs://s3adlogs/ad1.log")
	//assert.Nil(t, err)
	//err = service.Delete(object)
	//assert.Nil(t, err)

}
