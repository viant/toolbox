package gcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/kms"
	"github.com/viant/toolbox/storage"
	"github.com/viant/toolbox/url"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
	"io/ioutil"
)

type KmsService interface {
	Encrypt(ctx context.Context, key string, value string) (string, error)
	Decrypt(ctx context.Context, key string, value string) (string, error)
}

func (k *kmsService) Encrypt(ctx context.Context, key string, plainText string) (string, error) {
	kms, err := cloudkms.NewService(ctx, option.WithScopes(cloudkms.CloudPlatformScope, cloudkms.CloudkmsScope))
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to create kms server for key %v", key))
	}
	service := cloudkms.NewProjectsLocationsKeyRingsCryptoKeysService(kms)

	response, err := service.Encrypt(key, &cloudkms.EncryptRequest{Plaintext: plainText}).Context(ctx).Do()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to encrypt with key %v", key))
	}
	return response.Ciphertext, nil
}

func (k *kmsService) Decrypt(ctx context.Context, key string, plainText string) (string, error) {
	kms, err := cloudkms.NewService(ctx, option.WithScopes(cloudkms.CloudPlatformScope, cloudkms.CloudkmsScope))
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to create kms server for key %v", key))
	}
	service := cloudkms.NewProjectsLocationsKeyRingsCryptoKeysService(kms)
	response, err := service.Decrypt(key, &cloudkms.DecryptRequest{Ciphertext: plainText}).Context(ctx).Do()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to encrypt with key %v", key))
	}
	return response.Plaintext, nil
}

type kmsService struct{}

type service struct {
	KmsService
}

//New returns service
func New() kms.Service {
	return newService()
}

func newService() kms.Service {
	return &service{KmsService: &kmsService{}}
}

func (s *service) Decode(ctx context.Context, decryptRequest *kms.DecryptRequest, factory toolbox.DecoderFactory, target interface{}) error {
	response, err := s.Decrypt(ctx, decryptRequest)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(response.Data)
	return factory.Create(reader).Decode(target)
}

func (s *service) Encrypt(ctx context.Context, request *kms.EncryptRequest) (*kms.EncryptResponse, error) {

	if request.URL != "" {
		data, err := getDataFromURL(request.URL)
		if err != nil {
			return nil, err
		}
		if data == nil || len(data) == 0 {
			return nil, fmt.Errorf("data empty in the encrypt")
		}
		request.Data = data

	}
	plainText := getBase64(request.Data)
	encryptedText, err := s.KmsService.Encrypt(ctx, request.Key, plainText)
	if err != nil {
		return nil, err
	}
	if encryptedText == "" {
		return nil, fmt.Errorf("encryptedText was empty")
	}

	if request.TargetURL != "" {
		err = upload(request.TargetURL, encryptedText)
		if err != nil {
			return nil, err
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

func (s *service) Decrypt(ctx context.Context, request *kms.DecryptRequest) (*kms.DecryptResponse, error) {
	if request.URL != "" {
		resource := url.NewResource(request.URL)
		base64Text, err := resource.DownloadBase64()
		if err != nil {
			return nil, err
		}
		request.Data = []byte(base64Text)
	} else if len(request.Data) > 0 {
		base64Text := getBase64(request.Data)
		request.Data = []byte(base64Text)
	}
	plainText := string(request.Data)
	text, err := s.KmsService.Decrypt(ctx, request.Key, plainText)
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, fmt.Errorf("no text in the decrypt")
	}

	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to base64 decode text %v", text))
	}
	decryptResponse := &kms.DecryptResponse{
		Data: data,
		Text: text,
	}
	return decryptResponse, nil
}

func getBase64(data []byte) string {
	plainText := string(data)
	isBase64 := false
	if _, err := base64.StdEncoding.DecodeString(string(data)); err == nil {
		isBase64 = true
	}

	if !isBase64 {
		plainText = base64.StdEncoding.EncodeToString(data)
	}
	return plainText
}

func upload(targetURL string, encryptedText string) error {
	storageService, err := storage.NewServiceForURL(targetURL, "")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get storage for url %v", targetURL))
	}
	return storageService.Upload(targetURL, bytes.NewReader([]byte(encryptedText)))
}

func getDataFromURL(URL string) ([]byte, error) {
	storageService, err := storage.NewServiceForURL(URL, "")
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create storage for url: %v", URL))
	}
	reader, err := storageService.DownloadWithURL(URL)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to download url: %v", URL))
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read data from %v", URL))
	}
	return data, nil

}
