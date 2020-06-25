package storage_test

import "testing"

func TestNewHttpStorageService(t *testing.T) {

	// Deprecated use gith	ub/viant/afs/http
	//{
	//	service, err := storage.NewServiceForURL("https://github.com/viant/", credentialFile)
	//	assert.Nil(t, err)
	//	assert.NotNil(t, service)
	//
	//	objects, err := service.List("https://github.com/viant/")
	//	assert.True(t, len(objects) > 0)
	//	reader, err := service.Download(objects[0])
	//	if assert.Nil(t, err) {
	//		defer reader.Close()
	//	}
	//	content, err := ioutil.ReadAll(reader)
	//	assert.Nil(t, err)
	//	assert.True(t, len(content) > 0)
	//
	//}

}
