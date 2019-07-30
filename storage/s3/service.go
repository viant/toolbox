package s3

import (
	"bytes"
	"fmt"
	"github.com/viant/toolbox/cred"
	"io"
	"net/url"
	"strings"

	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	"os"
)

var defaultTime = time.Time{}

type service struct {
	config *cred.Config
}

func listFolders(client *s3.S3, url *url.URL, result *[]storage.Object) error {
	folderRequest := &s3.ListObjectsInput{
		Bucket:    aws.String(url.Host),
		Prefix:    aws.String(url.Path[1:]),
		Delimiter: aws.String("/"),
	}
	prefixes := make([]*s3.CommonPrefix, 0)
	err := client.ListObjectsPages(folderRequest,
		func(page *s3.ListObjectsOutput, lastPage bool) bool {
			prefixes = append(prefixes, page.CommonPrefixes...)
			return len(page.CommonPrefixes) > 0
		})

	if err != nil {
		if strings.Contains(err.Error(), "BucketRegionError") {
			return nil
		}
		return err
	}
	for _, prefix := range prefixes {
		pathURL := "s3://" + url.Host + "/" + *prefix.Prefix
		var _, name = toolbox.URLSplit(pathURL)
		var fileMode, _ = storage.NewFileMode("drw-rw-rw-")
		var fileInfo = storage.NewFileInfo(name, 102, fileMode, defaultTime, fileMode.IsDir())
		var object = newStorageObject(pathURL, prefix, fileInfo)
		*result = append(*result, object)
	}
	return nil
}

func listContent(client *s3.S3, parsedURL *url.URL, result *[]storage.Object) error {
	var path = parsedURL.Path

	folderRequest := &s3.ListObjectsInput{
		Bucket:    aws.String(parsedURL.Host),
		Delimiter: aws.String("/"),
	}
	if len(path) > 0 {
		folderRequest.Prefix = aws.String(parsedURL.Path[1:])
	}
	contents := make([]*s3.Object, 0)
	err := client.ListObjectsPages(folderRequest,
		func(page *s3.ListObjectsOutput, lastPage bool) bool {
			contents = append(contents, page.Contents...)
			return len(page.Contents) > 0
		})

	if err != nil {
		if strings.Contains(err.Error(), "BucketRegionError") {
			return nil
		}
		return err
	}
	for _, content := range contents {
		objectURL := "s3://" + parsedURL.Host + "/" + *content.Key
		var _, name = toolbox.URLSplit(objectURL)
		var fileMode, _ = storage.NewFileMode("-rw-rw-rw-")
		var fileInfo = storage.NewFileInfo(name, *content.Size, fileMode, *content.LastModified, fileMode.IsDir())
		var object = newStorageObject(objectURL, content, fileInfo)
		*result = append(*result, object)
	}
	return nil
}

func (s *service) getAwsConfig() (*aws.Config, error) {
	if s.config.Secret == "" {
		return aws.NewConfig().WithRegion(s.config.Region), nil
	}
	awsCredentials := credentials.NewStaticCredentials(s.config.Key, s.config.Secret, s.config.Token)
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
		return nil, fmt.Errorf("failed to parse : %v", err)
	}
	config, err := s.getAwsConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get aws config: %v", err)
	}
	client := s3.New(session.New(), config)
	err = listFolders(client, u, &result)
	if err == nil {
		err = listContent(client, u, &result)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get list content: %v", err)
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
		return nil, fmt.Errorf("Not found %v", URL)
	}
	return objects[0], nil
}

func (s *service) Download(object storage.Object) (io.ReadCloser, error) {
	u, err := url.Parse(object.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse : %v", err)
	}
	config, err := s.getAwsConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get aws config: %v", err)
	}
	downloader := s3manager.NewDownloader(session.New(config))
	target := &s3.Object{}
	_ = object.Unwrap(&target)
	writer := toolbox.NewByteWriterAt()
	_, err = downloader.Download(writer,
		&s3.GetObjectInput{
			Bucket: aws.String(u.Host),
			Key:    aws.String(*target.Key),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to download: %v", err)
	}
	return ioutil.NopCloser(bytes.NewReader(writer.Buffer)), nil

}

func (s *service) Upload(URL string, reader io.Reader) error {
	return s.UploadWithMode(URL, storage.DefaultFileMode, reader)
}

func (s *service) UploadWithMode(URL string, mode os.FileMode, reader io.Reader) error {
	err := s.uploadContent(URL, reader)
	if toolbox.IsNotFoundError(err) {
		config, err := s.getAwsConfig()
		if err != nil {
			return err
		}
		if parserURL, err := url.Parse(URL); err == nil {
			client := s3.New(session.New(config))
			if _, err := client.CreateBucket(&s3.CreateBucketInput{
				Bucket: &parserURL.Host,
			}); err != nil {
				return err
			}
		}
		return s.uploadContent(URL, reader)
	}
	return err
}

func (s *service) uploadContent(URL string, reader io.Reader) error {
	parsedURL, err := url.Parse(URL)
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
		Bucket: aws.String(parsedURL.Host),
		Key:    aws.String(parsedURL.Path),
	})
	if err != nil {
		return toolbox.ReclassifyNotFoundIfMatched(err, URL)
	}
	return nil
}

func (s *service) Delete(object storage.Object) error {
	parsedURL, err := url.Parse(object.URL())
	if err != nil {
		return err
	}

	if object.IsFolder() {
		var objects = []storage.Object{}
		objects, err = s.List(object.URL())
		if err != nil {
			return err
		}
		for _, object := range objects {
			if err := s.Delete(object); err != nil {
				return err
			}
		}
		return nil
	}

	target := &s3.Object{}
	object.Unwrap(&target)
	request := &s3.DeleteObjectInput{
		Bucket: aws.String(parsedURL.Host),
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

//DownloadWithURL downloads content for passed in object URL
func (s *service) DownloadWithURL(URL string) (io.ReadCloser, error) {
	object, err := s.StorageObject(URL)
	if err != nil {
		return nil, err
	}
	return s.Download(object)
}

func (s *service) Close() error {
	return nil
}

//NewService creates a new aws storage service
func NewService(config *cred.Config) storage.Service {
	return &service{config: config}
}
