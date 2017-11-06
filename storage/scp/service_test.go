package scp_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/viant/toolbox/storage/scp"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"github.com/viant/toolbox/storage"
)

func TestService_List(t *testing.T) {

	service := scp.NewService(nil)

	assert.NotNil(t, service)

	dir, home := path.Split(os.Getenv("HOME"))
	objects, err := service.List("scp://127.0.0.1/" + dir)
	assert.Nil(t, err)
	for _, object := range objects {
		if strings.HasSuffix(object.URL(), home) {
			assert.True(t, object.IsFolder())
		}
	}

}

func TestService_Delete(t *testing.T) {

	service := scp.NewService(nil)

	assert.NotNil(t, service)

	err := service.Upload("scp://127.0.0.1//tmp/file.txt", strings.NewReader("this is test"))
	assert.Nil(t, err)

	objects, err := service.List("scp://127.0.0.1/tmp/")
	assert.Nil(t, err)
	assert.True(t, len(objects) > 1)

	object, err := service.StorageObject("scp://127.0.0.1//tmp/file.txt")
	assert.Nil(t, err)

	reader, err := service.Download(object)
	if assert.Nil(t, err) {
		content, err := ioutil.ReadAll(reader)
		assert.Nil(t, err)
		assert.Equal(t, "this is test", string(content))
	}

	err = service.Delete(object)
	assert.Nil(t, err)

}

func Test_ExtractFileInfo(t *testing.T) {

	{
		var object storage.Object = scp.ExtractFileInfo("-rw-r--r--  1 awitas  1742120565   414 Jun  8 14:14:08 2017 f.pub", false)
		assert.Equal(t, "/f.pub", object.URL())
		assert.Equal(t, int64(1496931248), object.LastModified().Unix())
		assert.Equal(t, int64(414), object.Size())
		assert.Equal(t, true, object.IsContent())
	}


	var object storage.Object = scp.ExtractFileInfo("-rw-r--r-- 1 awitas awitas 2002 2017-11-04 22:29:33.363458941 +0000 aerospikeciads_aerospike.conf", true)
	assert.Equal(t, "/aerospikeciads_aerospike.conf", object.URL())
	assert.Equal(t, int64(1509834573), object.LastModified().Unix())
	assert.Equal(t, int64(2002), object.Size())
	assert.Equal(t, true, object.IsContent())

}