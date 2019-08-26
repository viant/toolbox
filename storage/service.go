// Package storage define abstract storage operation
// Deprecated - please use https://github.com/viant/afs API instead
// This package is frozen and no new functionality will be added, and future removal takes place.
package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
)

var DefaultFileMode os.FileMode = 0755

//Service represents abstract way to accessing local or remote storage
type Service interface {
	//List returns a list of object for supplied url
	List(URL string) ([]Object, error)
	//Exists returns true if resource exists
	Exists(URL string) (bool, error)
	//Object returns a Object for supplied url
	StorageObject(URL string) (Object, error)
	//Download returns reader for downloaded storage object
	Download(object Object) (io.ReadCloser, error)
	//DownloadWithURL returns reader for downloaded URL object
	DownloadWithURL(URL string) (io.ReadCloser, error)
	//Upload uploads provided reader content for supplied storage object.
	Upload(URL string, reader io.Reader) error
	//Upload uploads provided reader content for supplied storage object.
	UploadWithMode(URL string, mode os.FileMode, reader io.Reader) error
	//Delete removes passed in storage object
	Delete(object Object) error
	//Register register schema with provided service
	Register(schema string, service Service) error
	//Closes storage service
	Close() error
}

type storageService struct {
	registry map[string]Service
}

func (s *storageService) getServiceForSchema(URL string) (Service, error) {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	if result, found := s.registry[parsedUrl.Scheme]; found {
		return result, nil
	}

	return nil, fmt.Errorf("failed to lookup url schema %v in %v", parsedUrl.Scheme, URL)
}

//List lists all object for passed in URL
func (s *storageService) List(URL string) ([]Object, error) {
	service, err := s.getServiceForSchema(URL)
	if err != nil {
		return nil, err
	}
	return service.List(URL)
}

//Exists returns true if resource exists
func (s *storageService) Exists(URL string) (bool, error) {
	service, err := s.getServiceForSchema(URL)
	if err != nil {
		return false, err
	}
	return service.Exists(URL)
}

//StorageObject returns storage object for provided URL
func (s *storageService) StorageObject(URL string) (Object, error) {
	service, err := s.getServiceForSchema(URL)
	if err != nil {
		return nil, err
	}
	return service.StorageObject(URL)
}

//Download downloads content for passed in object
func (s *storageService) Download(object Object) (io.ReadCloser, error) {
	service, err := s.getServiceForSchema(object.URL())
	if err != nil {
		return nil, err
	}
	return service.Download(object)
}

//DownloadWithURL downloads content for passed in object URL
func (s *storageService) DownloadWithURL(URL string) (io.ReadCloser, error) {
	object, err := s.StorageObject(URL)
	if err != nil {
		return nil, err
	}
	return s.Download(object)
}

//Uploads content for passed in URL
func (s *storageService) Upload(URL string, reader io.Reader) error {
	service, err := s.getServiceForSchema(URL)
	if err != nil {
		return err
	}
	return service.UploadWithMode(URL, 0644, reader)
}

//Uploads content for passed in URL
func (s *storageService) UploadWithMode(URL string, mode os.FileMode, reader io.Reader) error {
	service, err := s.getServiceForSchema(URL)
	if err != nil {
		return err
	}
	return service.UploadWithMode(URL, mode, reader)
}

//Delete remove storage object
func (s *storageService) Delete(object Object) error {
	service, err := s.getServiceForSchema(object.URL())
	if err != nil {
		return err
	}
	return service.Delete(object)
}

//Close closes resources
func (s *storageService) Close() error {
	for _, service := range s.registry {
		err := service.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

//Register register storage schema
func (s *storageService) Register(schema string, service Service) error {
	s.registry[schema] = service
	return nil
}

//NewService creates a new storage service
func NewService() Service {
	var result = &storageService{
		registry: make(map[string]Service),
	}
	_ = result.Register("file", &fileStorageService{})
	_ = result.Register("mem", NewMemoryService())
	return result
}

//NewServiceForURL creates a new storage service for provided URL scheme and optional credential file
func NewServiceForURL(URL, credentials string) (Service, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	service := NewService()
	provider := Registry().Get(parsedURL.Scheme)

	if provider != nil {
		if len(credentials) > 0 {
			credentials = strings.Replace(credentials, "${env.HOME}", os.Getenv("HOME"), 1)
			if strings.HasPrefix(credentials, "~") {
				credentials = strings.Replace(credentials, "~", os.Getenv("HOME"), 1)
			}
		}
		serviceForScheme, err := provider(credentials)
		if err != nil {
			return nil, fmt.Errorf("failed lookup service for %v: %v", parsedURL.Scheme, err)
		}
		err = service.Register(parsedURL.Scheme, serviceForScheme)
		if err != nil {
			return nil, err
		}
	} else if parsedURL.Scheme != "file" {
		return nil, fmt.Errorf("unsupported scheme %v", URL)
	}
	return service, nil
}

//Download returns a download reader for supplied URL
func Download(service Service, URL string) (io.ReadCloser, error) {
	object, err := service.StorageObject(URL)
	if err != nil {
		return nil, err
	}
	return service.Download(object)
}

//DownloadText returns a text for supplied URL
func DownloadText(service Service, URL string) (string, error) {
	reader, err := Download(service, URL)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
