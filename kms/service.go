package kms

import (
	"context"
	"github.com/viant/toolbox"
)

type Service interface {
	Encrypt(context.Context, *EncryptRequest) (*EncryptResponse, error)

	Decrypt(context.Context, *DecryptRequest) (*DecryptResponse, error)

	Decode(ctx context.Context, decryptRequest *DecryptRequest, factory toolbox.DecoderFactory, target interface{}) error
}
