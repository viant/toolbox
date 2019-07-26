package aws

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	akms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
	"strings"

	"github.com/viant/toolbox"
	"github.com/viant/toolbox/kms"
)

type service struct {
	*ssm.SSM
	*akms.KMS
}

func (s *service) Encrypt(ctx context.Context, request *kms.EncryptRequest) (*kms.EncryptResponse, error) {
	err := request.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "invalid encrypt request")
	}
	if request.Parameter == "" {
		return nil, errors.New("parameter was empty")
	}
	response := &kms.EncryptResponse{}
	err = s.putParameters(request.Key, request.Parameter, string(request.Data))
	if err == nil {
		parameter, err := s.getParameters(request.Parameter, false)
		if err != nil {
			return nil, err
		}
		response.EncryptedText = *parameter.Value
		response.EncryptedData = []byte(response.EncryptedText)
	}
	return response, err
}

func (s *service) Decrypt(ctx context.Context, request *kms.DecryptRequest) (*kms.DecryptResponse, error) {
	err := request.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "invalid encrypt request")
	}
	if request.Parameter == "" {
		return nil, errors.New("parameter was empty")
	}
	response := &kms.DecryptResponse{}
	parameter, err := s.getParameters(request.Parameter, true)
	if err != nil {
		return nil, err
	}
	response.Text = *parameter.Value
	response.Data = []byte(response.Text)
	return response, nil
}

func (s *service) Decode(ctx context.Context, decryptRequest *kms.DecryptRequest, factory toolbox.DecoderFactory, target interface{}) error {
	response, err := s.Decrypt(ctx, decryptRequest)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(response.Data)
	return factory.Create(reader).Decode(target)
}

func (s *service) putParameters(keyOrAlias, name, value string) error {
	targetKeyID, err := s.getKeyByAlias(keyOrAlias)
	if err != nil {
		return err
	}
	_, err = s.PutParameter(&ssm.PutParameterInput{
		Name:  aws.String(name),
		KeyId: &targetKeyID,
		Value: &value,
	})
	return err
}

func (s *service) getKeyByAlias(keyOrAlias string) (string, error) {
	if strings.Count(keyOrAlias, ":") > 0 {
		return keyOrAlias, nil
	}
	var nextMarker *string
	for {
		output, err := s.ListAliases(&akms.ListAliasesInput{
			Marker: nextMarker,
		})
		if err != nil {
			return "", err
		}
		if len(output.Aliases) == 0 {
			break
		}
		for _, candidate := range output.Aliases {
			if *candidate.AliasName == keyOrAlias {
				return *candidate.TargetKeyId, nil
			}
		}
		nextMarker = output.NextMarker
		if nextMarker == nil {
			break
		}
	}
	return "", fmt.Errorf("key for alias %v no found", keyOrAlias)
}

func (s *service) getParameters(name string, withDecryption bool) (*ssm.Parameter, error) {
	output, err := s.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return nil, err
	}
	return output.Parameter, nil
}

//New returns new kms service
func New() (kms.Service, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &service{
		SSM: ssm.New(sess),
		KMS: akms.New(sess),
	}, nil
}
