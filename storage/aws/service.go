package aws

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	"io"
	"net/url"
	"strings"
)

type service struct {
	config *Config
}

func listFolders(client *s3.S3, url *url.URL, result *[]storage.Object) error {
	folderRequest := &s3.ListObjectsInput{
		Bucket:    aws.String(url.Host),
		Prefix:    aws.String(url.Path[1:]),
		Delimiter: aws.String("/"),
	}
	resp, err := client.ListObjects(folderRequest)
	if err != nil {
		if strings.Contains(err.Error(), "BucketRegionError") {
			return nil
		}
		return err
	}
	for _, prefix := range resp.CommonPrefixes {
		path := "s3://" + url.Host + "/" + *prefix.Prefix
		var object = newObject(path, storage.StorageObjectFolderType, prefix, nil, 0)
		*result = append(*result, object)
	}
	return nil
}

func listContent(client *s3.S3, parsedUrl *url.URL, result *[]storage.Object) error {
	var path = parsedUrl.Path

	folderRequest := &s3.ListObjectsInput{
		Bucket:    aws.String(parsedUrl.Host),
		Delimiter: aws.String("/"),
	}
	if len(path) > 0 {
		folderRequest.Prefix = aws.String(parsedUrl.Path[1:])
	}
	resp, err := client.ListObjects(folderRequest)
	if err != nil {
		if strings.Contains(err.Error(), "BucketRegionError") {
			return nil
		}
		return err
	}
	for _, content := range resp.Contents {
		path := "s3://" + parsedUrl.Host + "/" + *content.Key

		var object = newObject(path, storage.StorageObjectContentType, content, content.LastModified, *content.Size)

		*result = append(*result, object)
	}
	return nil
}

func (s *service) getAwsConfig() (*aws.Config, error) {
	awsCredentials := credentials.NewStaticCredentials(s.config.Key, s.config.Secrect, s.config.Token)
	_, err := awsCredentials.Get()
	if err != nil {
		return nil, fmt.Errorf("bad credentials: %s", err)
	}
	return aws.NewConfig().WithRegion(s.config.Region).WithCredentials(awsCredentials), nil
}

func (s *service) List(URL string) ([]storage.Object, error) {
	var result = make([]storage.Object, 0)
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	config, err := s.getAwsConfig()
	if err != nil {
		return nil, err
	}
	client := s3.New(session.New(), config)
	err = listFolders(client, u, &result)
	if err == nil {
		err = listContent(client, u, &result)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) Exists(URL string) (bool, error) {
	objects, err := s.List(URL)
	if err != nil {
		return false, err
	}
	return len(objects) > 0, nil
}

func (s *service) StorageObject(URL string) (storage.Object, error) {
	objects, err := s.List(URL)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("No found %v", URL)
	}
	return objects[0], nil
}

func (s *service) Download(object storage.Object) (io.Reader, error) {
	u, err := url.Parse(object.URL())
	if err != nil {
		return nil, err
	}

	config, err := s.getAwsConfig()
	if err != nil {
		return nil, err
	}
	downloader := s3manager.NewDownloader(session.New(config))
	target := &s3.Object{}
	object.Unwrap(&target)
	writer := toolbox.NewByteWriterAt()
	_, err = downloader.Download(writer,
		&s3.GetObjectInput{
			Bucket: aws.String(u.Host),
			Key:    aws.String(*target.Key),
		})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(writer.Buffer), nil

}

func (s *service) Upload(URL string, reader io.Reader) error {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return err
	}
	config, err := s.getAwsConfig()
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(session.New(config))
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: aws.String(parsedUrl.Host),
		Key:    aws.String(parsedUrl.Path),
	})
	if err != nil {
		return fmt.Errorf("Failed to upload %v", err)
	}
	return nil
}

func (s *service) Delete(object storage.Object) error {
	parsedUrl, err := url.Parse(object.URL())
	if err != nil {
		return err
	}
	target := &s3.Object{}
	object.Unwrap(&target)
	request := &s3.DeleteObjectInput{
		Bucket: aws.String(parsedUrl.Host),
		Key:    target.Key,
	}
	config, err := s.getAwsConfig()
	if err != nil {
		return err
	}
	client := s3.New(session.New(), config)
	client.DeleteObject(request)
	return nil
}

func (s *service) Register(schema string, service storage.Service) error {
	return fmt.Errorf("Unsupported")
}

//NewService creates a new aws storage service
func NewService(config *Config) storage.Service {
	return &service{config: config}
}
