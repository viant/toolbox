package toolbox

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"time"
)

//StorageService represents abstract way to accessing local or remote storage
type fileStorageService struct{}

func openFileFromUrl(URL string) (*os.File, error) {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	if parsedUrl.Scheme != "file" {
		return nil, fmt.Errorf("Invalid schema, expected file but had: %v", parsedUrl.Scheme)
	}
	return os.Open(parsedUrl.Path)
}

//List returns a list of object for supplied url
func (s *fileStorageService) List(URL string) ([]StorageObject, error) {
	file, err := openFileFromUrl(URL)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	files, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}
	var result = make([]StorageObject, 0)
	for _, fileInfo := range files {
		fileUrl := URL + "/" + fileInfo.Name()
		objectType := StorageObjectContentType
		if fileInfo.IsDir() {
			objectType = StorageObjectFolderType
		}
		modTime := fileInfo.ModTime()
		result = append(result, newFileObject(fileUrl, objectType, fileInfo, &modTime, fileInfo.Size()))
	}
	return result, nil
}

//StorageObject returns a StorageObject for supplied url
func (s *fileStorageService) StorageObject(URL string) (StorageObject, error) {
	file, err := openFileFromUrl(URL)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, err := os.Stat(file.Name())
	if err != nil {
		return nil, err
	}
	objectType := 0
	switch mode := fileInfo.Mode(); {

	case mode.IsDir():
		// do directory stuff
		objectType = StorageObjectFolderType
	case mode.IsRegular():
		objectType = StorageObjectContentType
	}
	modTime := fileInfo.ModTime()
	return newFileObject(URL, objectType, fileInfo, &modTime, fileInfo.Size()), nil
}

//Download returns reader for downloaded storage object
func (s *fileStorageService) Download(object StorageObject) (io.Reader, error) {
	reader, _, err := OpenReaderFromURL(object.URL())
	return reader, err
}

//Upload uploads provided reader content for supplied url.
func (s *fileStorageService) Upload(URL string, reader io.Reader) error {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if parsedUrl.Scheme != "file" {
		return fmt.Errorf("Invalid schema, expected file but had: %v", parsedUrl.Scheme)
	}

	parentDir, _ := path.Split(parsedUrl.Path)
	err = CreateDirIfNotExist(parentDir)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(parsedUrl.Path, data, fileMode)
}

func (s *fileStorageService) Register(schema string, service StorageService) error {
	return errors.New("unsupported")
}

//Delete removes passed in storage object
func (s *fileStorageService) Delete(object StorageObject) error {
	var fileInfo *os.FileInfo
	object.Unwrap(fileInfo)
	return os.Remove((*fileInfo).Name())
}

type fileStorageObject struct {
	*AbstractStorageObject
}

func (o *fileStorageObject) Unwrap(target interface{}) error {
	if fileInfo, casted := target.(**os.FileInfo); casted {
		source, ok := o.Source.(*os.FileInfo)
		if !ok {
			return fmt.Errorf("Failed to case %T into %T", o.Source, target)
		}
		*fileInfo = source
		return nil
	}
	return fmt.Errorf("unsuported target %T", target)
}

func newFileObject(url string, objectType int, source interface{}, lastModified *time.Time, size int64) StorageObject {
	abstract := NewAbstractStorageObject(url, source, objectType, lastModified, size)
	result := &fileStorageObject{
		AbstractStorageObject: abstract,
	}
	result.AbstractStorageObject.StorageObject = result
	return result
}
