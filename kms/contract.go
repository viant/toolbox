package kms

import "errors"

type Resource struct {
	URL       string
	Parameter string
	Data      []byte
}

type EncryptRequest struct {
	Key string
	*Resource
	TargetURL string
}

type EncryptResponse struct {
	EncryptedData []byte
	EncryptedText string
}

type DecryptRequest struct {
	Key string
	*Resource
}

type DecryptResponse struct {
	Data []byte
	Text string
}

func (r *EncryptRequest) Validate() error {
	if r.Key == "" {
		return errors.New("key was empty")
	}
	if r.Resource == nil {
		return errors.New("nothing to encrypt")
	}
	return nil
}

func (r *DecryptRequest) Validate() error {
	if r.Key == "" {
		return errors.New("key was empty")
	}
	if r.Resource == nil {
		return errors.New("nothing to decrypt")
	}
	return nil

}
