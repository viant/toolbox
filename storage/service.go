package storage

import (
	"fmt"
	"io"
	"net/url"
)

//Service represents abstract way to accessing local or remote storage
type Service interface {
	//List returns a list of object for supplied url
	List(URL string) ([]Object, error)

	//Exists returns true if resource exists
	Exists(URL string) (bool, error)

	//Object returns a Object for supplied url
	StorageObject(URL string) (Object, error)

	//Download returns reader for downloaded storage object
	Download(object Object) (io.Reader, error)

	//Upload uploads provided reader content for supplied storage object.
	Upload(URL string, reader io.Reader) error

	//Delete removes passed in storage object
	Delete(object Object) error

	//Register register schema with provided service
	Register(schema string, service Service) error
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

	return nil, fmt.Errorf("Failed to lookup url schema %v in %v", parsedUrl.Scheme, URL)
}

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

func (s *storageService) StorageObject(URL string) (Object, error) {
	service, err := s.getServiceForSchema(URL)
	if err != nil {
		return nil, err
	}
	return service.StorageObject(URL)
}

func (s *storageService) Download(object Object) (io.Reader, error) {
	service, err := s.getServiceForSchema(object.URL())
	if err != nil {
		return nil, err
	}
	return service.Download(object)
}

func (s *storageService) Upload(URL string, reader io.Reader) error {
	service, err := s.getServiceForSchema(URL)
	if err != nil {
		return err
	}
	return service.Upload(URL, reader)
}

func (s *storageService) Delete(object Object) error {
	service, err := s.getServiceForSchema(object.URL())
	if err != nil {
		return err
	}
	return service.Delete(object)
}

func (s *storageService) Register(schema string, service Service) error {
	s.registry[schema] = service
	return nil
}

func NewService() Service {
	var result = &storageService{
		registry: make(map[string]Service),
	}
	result.Register("file", &fileStorageService{})
	return result
}
