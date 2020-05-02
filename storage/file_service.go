package storage

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/toolbox"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
)

var fileMode os.FileMode = 0644
var execFileMode os.FileMode = 0755

//Service represents abstract way to accessing local or remote storage
type fileStorageService struct{}

//List returns a list of object for supplied url
func (s *fileStorageService) List(URL string) ([]Object, error) {
	file, err := toolbox.OpenFile(URL)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err == nil {
		if !stat.IsDir() {
			return []Object{
				newFileObject(URL, stat),
			}, nil
		}
	}
	files, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}
	var result = make([]Object, 0)
	result = append(result, newFileObject(URL, stat))

	var parsedURL, _ = url.Parse(URL)
	for _, fileInfo := range files {
		var fileName = fileInfo.Name()
		if parsedURL != nil {
			fileName = strings.Replace(fileName, parsedURL.Path, "", 1)
		}

		fileURL := toolbox.URLPathJoin(URL, fileName)
		result = append(result, newFileObject(fileURL, fileInfo))
	}
	return result, nil
}

//Exists returns true if resource exists
func (s *fileStorageService) Exists(URL string) (bool, error) {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return false, err
	}
	if parsedUrl.Scheme != "file" {
		return false, fmt.Errorf("invalid schema, expected file but had: %v", parsedUrl.Scheme)
	}
	return toolbox.FileExists(parsedUrl.Path), nil
}

func (s *fileStorageService) Close() error {
	return nil
}

//Object returns a Object for supplied url
func (s *fileStorageService) StorageObject(URL string) (Object, error) {
	file, err := toolbox.OpenFile(URL)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, err := os.Stat(file.Name())
	if err != nil {
		return nil, err
	}
	return newFileObject(URL, fileInfo), nil
}

//Download returns reader for downloaded storage object
func (s *fileStorageService) Download(object Object) (io.ReadCloser, error) {
	return toolbox.OpenFile(object.URL())
}

//DownloadWithURL downloads content for passed in object URL
func (s *fileStorageService) DownloadWithURL(URL string) (io.ReadCloser, error) {
	object, err := s.StorageObject(URL)
	if err != nil {
		return nil, err
	}
	return s.Download(object)
}

func (s *fileStorageService) Upload(URL string, reader io.Reader) error {
	return s.UploadWithMode(URL, DefaultFileMode, reader)
}

//Upload uploads provided reader content for supplied url.
func (s *fileStorageService) UploadWithMode(URL string, mode os.FileMode, reader io.Reader) error {
	if mode == 0 {
		mode = DefaultFileMode
	}
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if parsedUrl.Scheme != "file" {
		return fmt.Errorf("Invalid schema, expected file but had: %v", parsedUrl.Scheme)
	}

	parentDir, _ := path.Split(parsedUrl.Path)

	err = toolbox.CreateDirIfNotExist(parentDir)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(parsedUrl.Path, data, mode)
}

func (s *fileStorageService) Register(schema string, service Service) error {
	return errors.New("unsupported")
}

//Delete removes passed in storage object
func (s *fileStorageService) Delete(object Object) error {

	if object.IsFolder() {
		objects, err := s.List(object.URL())
		if err != nil {
			return err
		}
		for _, listedObject := range objects {
			if listedObject.URL() == object.URL() {
				continue
			}
			if err := s.Delete(listedObject); err != nil {
				return err
			}
		}

	}
	fileName := toolbox.Filename(object.URL())
	return os.Remove(fileName)
}

type fileStorageObject struct {
	*AbstractObject
}

func NewFileStorage() Service {
	return &fileStorageService{}
}

func (o *fileStorageObject) Unwrap(target interface{}) error {
	if fileInfo, casted := target.(*os.FileInfo); casted {
		source, ok := o.Source.(os.FileInfo)
		if !ok {
			return fmt.Errorf("failed to cast %T into %T", o.Source, target)
		}
		*fileInfo = source
		return nil
	}

	return fmt.Errorf("unsuported target %T", target)
}

func (o *fileStorageObject) FileInfo() os.FileInfo {
	if source, ok := o.Source.(os.FileInfo); ok {
		return source
	}
	return nil
}

func newFileObject(url string, fileInfo os.FileInfo) Object {
	abstract := NewAbstractStorageObject(url, fileInfo, fileInfo)
	result := &fileStorageObject{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}

const FileProviderSchema = "file"

func init() {
	Registry().Registry[FileProviderSchema] = fileServiceProvider

}

func fileServiceProvider(credentialFile string) (service Service, err error) {
	return NewFileStorage(), nil
}
