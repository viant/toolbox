package gcp

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/kms"
	"github.com/viant/toolbox/url"
	"golang.org/x/net/context"
	"strings"
	"testing"
)

type MyTestConfig struct {
	Aaa string `json:",omitempty"`
	Bbb string `json:",omitempty"`
	Ccc string `json:",omitempty"`
}

func TestDecoder(t *testing.T) {

	// test decrypt read from non-base64 url
	{

		decryptRequest := kms.DecryptRequest{}
		decryptRequest.Resource = &kms.Resource{}
		decryptRequest.Resource.URL = url.NewResource("test/data/config.txt", "").URL
		decoderFactory := toolbox.NewJSONDecoderFactory()
		mytestConfig := MyTestConfig{}
		service := service{KmsService: &testKmsService{}}
		service.Decode(context.Background(), &decryptRequest, decoderFactory, &mytestConfig)
		assert.Equal(t, mytestConfig.Aaa, "Test1")
		assert.Equal(t, mytestConfig.Bbb, "test2")
		assert.Equal(t, mytestConfig.Ccc, "test3")

	}

	// test decrypt read from base64 url
	{

		decryptRequest := kms.DecryptRequest{}
		decryptRequest.Resource = &kms.Resource{}
		decryptRequest.Resource.URL = url.NewResource("test/data/config_base_64.txt", "").URL
		decoderFactory := toolbox.NewJSONDecoderFactory()
		mytestConfig := MyTestConfig{}
		service := service{KmsService: &testKmsService{}}
		service.Decode(context.Background(), &decryptRequest, decoderFactory, &mytestConfig)
		assert.Equal(t, mytestConfig.Aaa, "Test111")
		assert.Equal(t, mytestConfig.Bbb, "test222")
		assert.Equal(t, mytestConfig.Ccc, "test333")

	}

}

func TestEncrypt(t *testing.T) {
	//test base64 as input which from non-url
	{
		text := "path with?reserved+characters"

		request := kms.EncryptRequest{}
		encryptTextAfterBase64 := base64.StdEncoding.EncodeToString([]byte(text))
		request.Resource = &kms.Resource{Data: []byte(encryptTextAfterBase64)}
		service := service{KmsService: &testKmsService{}}
		response, err := service.Encrypt(context.Background(), &request)
		assert.Nil(t, err)
		assert.Equal(t, response.EncryptedText, encryptTextAfterBase64)
		assert.Equal(t, string(response.EncryptedData), text)

		decryptRequest := kms.DecryptRequest{}
		decryptTextAfterBase64 := base64.StdEncoding.EncodeToString(response.EncryptedData)
		decryptRequest.Resource = &kms.Resource{Data: []byte(decryptTextAfterBase64)}
		decryptResponse, err := service.Decrypt(context.Background(), &decryptRequest)
		assert.Nil(t, err)
		assert.Equal(t, decryptResponse.Text, decryptTextAfterBase64)
		assert.Equal(t, string(decryptResponse.Data), text)
	}

	//test non-base64 as input which from non-url
	{
		text := "path with?reserved+characters2"

		request := kms.EncryptRequest{}

		request.Resource = &kms.Resource{Data: []byte(text)}
		service := service{KmsService: &testKmsService{}}
		response, err := service.Encrypt(context.Background(), &request)
		assert.Nil(t, err)
		assert.Equal(t, response.EncryptedText, base64.StdEncoding.EncodeToString([]byte(text)))
		assert.Equal(t, string(response.EncryptedData), text)

		decryptRequest := kms.DecryptRequest{}
		decryptRequest.Resource = &kms.Resource{Data: response.EncryptedData}
		decryptResponse, err := service.Decrypt(context.Background(), &decryptRequest)
		assert.Nil(t, err)
		assert.Equal(t, decryptResponse.Text, base64.StdEncoding.EncodeToString([]byte(text)))
		assert.Equal(t, string(decryptResponse.Data), text)
	}

	//test base64 as input which from url
	{

		request := kms.EncryptRequest{}
		request.Resource = &kms.Resource{}
		request.Resource.URL = url.NewResource("test/data/test1.txt", "").URL
		request.TargetURL = url.NewResource("test/upload/upload1.txt").URL
		service := service{KmsService: &testKmsService{}}
		response, err := service.Encrypt(context.Background(), &request)
		assert.Nil(t, err)
		assert.Equal(t, response.EncryptedText, base64.StdEncoding.EncodeToString([]byte("This is a encrypt/decrypt test!!")))
		assert.Equal(t, string(response.EncryptedData), "This is a encrypt/decrypt test!!")

		decryptRequest := kms.DecryptRequest{}
		decryptRequest.Resource = &kms.Resource{}
		decryptRequest.Resource.URL = url.NewResource("test/data/test1.txt", "").URL
		decryptResponse, err := service.Decrypt(context.Background(), &decryptRequest)
		assert.Nil(t, err)
		assert.Equal(t, decryptResponse.Text, base64.StdEncoding.EncodeToString([]byte("This is a encrypt/decrypt test!!")))
		assert.Equal(t, string(decryptResponse.Data), "This is a encrypt/decrypt test!!")
	}

	//test non-base64 as input which from url
	{

		request := kms.EncryptRequest{}
		request.Resource = &kms.Resource{}
		request.Resource.URL = url.NewResource("test/data/test2.txt", "").URL
		service := service{KmsService: &testKmsService{}}
		response, err := service.Encrypt(context.Background(), &request)
		assert.Nil(t, err)
		assert.Equal(t, response.EncryptedText, base64.StdEncoding.EncodeToString([]byte("This is a encrypt/decrypt test no2 !!!")))
		assert.Equal(t, string(response.EncryptedData), "This is a encrypt/decrypt test no2 !!!")

		decryptRequest := kms.DecryptRequest{}
		decryptRequest.Resource = &kms.Resource{}
		decryptRequest.Resource.URL = url.NewResource("test/data/test2.txt", "").URL
		decryptResponse, err := service.Decrypt(context.Background(), &decryptRequest)
		assert.Nil(t, err)
		assert.Equal(t, decryptResponse.Text, base64.StdEncoding.EncodeToString([]byte("This is a encrypt/decrypt test no2 !!!")))
		assert.Equal(t, string(decryptResponse.Data), "This is a encrypt/decrypt test no2 !!!")
	}

	//test non-base64 as input for encrypt, based64 as input for decrypt
	{

		request := kms.EncryptRequest{}
		request.Resource = &kms.Resource{}
		request.Resource.URL = url.NewResource("test/data/test3.txt", "").URL
		service := service{KmsService: &testKmsService{}}
		response, err := service.Encrypt(context.Background(), &request)
		assert.Nil(t, err)
		assert.Equal(t, response.EncryptedText, base64.StdEncoding.EncodeToString([]byte("This is a encrypt/decrypt test no3 !!!@@@")))
		assert.Equal(t, string(response.EncryptedData), "This is a encrypt/decrypt test no3 !!!@@@")

		decryptRequest := kms.DecryptRequest{}
		decryptRequest.Resource = &kms.Resource{}
		decryptRequest.Resource.URL = url.NewResource("test/data/test3_3.txt", "").URL
		decryptResponse, err := service.Decrypt(context.Background(), &decryptRequest)
		assert.Nil(t, err)
		assert.Equal(t, decryptResponse.Text, base64.StdEncoding.EncodeToString([]byte("This is a encrypt/decrypt test no3 !!!@@@")))
		assert.Equal(t, string(decryptResponse.Data), "This is a encrypt/decrypt test no3 !!!@@@")
	}

}

type testKmsService struct {
}

func getPath(value string) string {
	processUrl := url.NewResource(value)
	processedPath := strings.Replace(processUrl.URL, "file:/", "", -1)
	return processedPath
}

func (testKmsService *testKmsService) Encrypt(ctx context.Context, key string, plainText string) (string, error) {

	return plainText, nil
}

func (testKmsService *testKmsService) Decrypt(ctx context.Context, key string, plainText string) (string, error) {
	return plainText, nil
}
