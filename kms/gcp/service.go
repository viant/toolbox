package gcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/kms"
	"github.com/viant/toolbox/storage"
	"google.golang.org/api/cloudkms/v1"
	"io/ioutil"
	"google.golang.org/api/option"

)



type KmsService interface {
	Encrypt(ctx context.Context,key string,value string)  (string, error)
	Decrypt(ctx context.Context,key string,value string)  (string, error)
}

func (k *kmsService) Encrypt(ctx context.Context,key string,plainText string) (string, error) {
	kms, err := cloudkms.NewService(ctx, option.WithScopes(cloudkms.CloudPlatformScope, cloudkms.CloudkmsScope))
	if err != nil {
		return "", err
	}

	if err != nil {
		return "",err
	}
	service := cloudkms.NewProjectsLocationsKeyRingsCryptoKeysService(kms)

	response, err := service.Encrypt(key, &cloudkms.EncryptRequest{Plaintext: plainText}).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return response.Ciphertext,nil
}



func (k *kmsService) Decrypt(ctx context.Context,key string,plainText string) (string, error) {
	kms, err := cloudkms.NewService(ctx, option.WithScopes(cloudkms.CloudPlatformScope, cloudkms.CloudkmsScope))
	if err != nil {
		return "", err
	}
	if err != nil {
		return "",err
	}
	service := cloudkms.NewProjectsLocationsKeyRingsCryptoKeysService(kms)
	response, err := service.Decrypt(key, &cloudkms.DecryptRequest{Ciphertext:plainText}).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return response.Plaintext,nil
}

type kmsService struct {

}

type service struct {
	KmsService
}


var srv kms.Service

//GetService returns service
func GetService() (kms.Service) {
	if srv != nil {
		return srv
	}
	return newService()
}

func newService() kms.Service {
	return &service{KmsService:&kmsService{}}
}

func (s *service) Decode(ctx context.Context,decryptRequest *kms.DecryptRequest, factory toolbox.DecoderFactory,target interface{}) error {
	response, err := s.Decrypt(ctx,decryptRequest)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(response.Data)
	return factory.Create(reader).Decode(target)
}

func (s *service) Encrypt(ctx context.Context,request *kms.EncryptRequest) (*kms.EncryptResponse, error) {
	if request.URL != "" {
		data, err := getDataFromURL(request.URL)
		if err != nil {
			return nil , err
		}
		if data == nil || len(data) == 0 {
			return nil,fmt.Errorf("data empty in the encrypt")
		}
		request.Data = data

	}
	plainText := getPlainText(request.Data,request.IsBase64)
	encryptedText,err := s.KmsService.Encrypt(ctx,request.Key,plainText)
	if err != nil{
		return nil,err
	}
	if  encryptedText == "" {
		return nil,fmt.Errorf("no encryptedText in the encrypt")
	}

    if request.TargetURL != "" {
	 	err = upload(request.TargetURL,encryptedText)
	 	if err != nil {
	 		fmt.Printf("error = %v\n",err)
	 		return nil,err
		}
	}
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return nil, err
	}
	return &kms.EncryptResponse{
		EncryptedData: encryptedData,
		EncryptedText: encryptedText,
	}, nil
}




func (s *service) Decrypt(ctx context.Context,request *kms.DecryptRequest) (*kms.DecryptResponse, error) {
	if request.URL != "" {
		data, err := getDataFromURL(request.URL)
		if err != nil {
			return nil , err
		}
		if data == nil || len(data) == 0 {
			return nil,fmt.Errorf("data empty in the decrypt")
		}
		request.Data = data
	}

	plainText := getPlainText(request.Data,request.IsBase64)

	text, err := s.KmsService.Decrypt(ctx, request.Key, plainText)
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, fmt.Errorf("no text in the decrypt")
	}

	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	decryptResponse := &kms.DecryptResponse{
		Data:     data,
		Text: text,
	}

	return decryptResponse, nil
}

func getPlainText(data []byte,isBase64 bool) string {
	plainText := string(data)
	if !isBase64 {
		plainText = base64.StdEncoding.EncodeToString(data)
	}
	return plainText
}

func upload(TargetURL string,encryptedText string) error {
	storageService, err:= storage.NewServiceForURL(TargetURL, "")
	if err != nil {
		fmt.Printf("err when upload =%v\n",err)
		return err
	}
	return storageService.Upload(TargetURL,bytes.NewReader([]byte(encryptedText)))
}

func  getDataFromURL(URL string) ([]byte, error) {
	storageService, err:= storage.NewServiceForURL(URL, "")
	if err != nil {
		return nil,err
	}
	reader, err := storageService.DownloadWithURL(URL)
	if err != nil {
		return nil,err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil,err
	}
	return data,nil

}

