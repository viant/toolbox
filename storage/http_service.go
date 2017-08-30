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
	"net/http"
	"strings"
	"path/filepath"
	"time"
	"bytes"
)

//PasswordCredential represents a password based credential
type PasswordCredential struct {
	Username string
	Password string
}

//httpStorageService represents basic http storage service (only limited listing and full download are supported)
type httpStorageService struct {
	Credential *PasswordCredential
}

func (s *httpStorageService) addCredentialToURLIfNeeded(URL string) string {
	if s.Credential == nil || s.Credential.Password == "" || s.Credential.Username == "" {
		return URL
	}
	prasedURL, err := url.Parse(URL)
	if err != nil {
		return URL
	}
	if prasedURL.User != nil  {
		return URL
	}
	return strings.Replace(URL, "://", fmt.Sprintf("://%v:%v@", s.Credential.Username, s.Credential.Password), 1)
}

//List returns a list of object for supplied url
func (s *httpStorageService) List(URL string) ([]Object, error) {
	listURL := s.addCredentialToURLIfNeeded(URL)
	response, err := http.Get(listURL)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	contentType := response.Header.Get("Content-Type")
	var result = make([]Object, 0)

	if response.Status != "200 OK" {
		return nil, fmt.Errorf("Invalid response code: %v", response.Status)
	}
	if strings.Contains(contentType, "text/html") {
		var linkContents = strings.Split(string(body), "href=\"")
		for i := 1; i < len(linkContents); i++ {
			var linkContent = linkContents[i]
			linkEndPosition := strings.Index(linkContent, "\"")
			if linkEndPosition != -1 {
				linkURL := string(linkContent[:linkEndPosition])
				if  linkURL == "" || strings.Contains(linkURL, ":") || strings.HasPrefix(linkURL, "#") || strings.HasPrefix(linkURL, "?") || strings.HasPrefix(linkURL, ".") || strings.HasPrefix(linkURL, "/") {
					continue
				}
				linkType := StorageObjectContentType
				if strings.HasSuffix(linkURL, "/") {
					linkType = StorageObjectFolderType
				}
				objectURL := toolbox.URLPathJoin(URL, linkURL)
				storageObject := newHttpFileObject(objectURL, linkType, nil, &now, 0)
				result = append(result, storageObject)
			}
		}
	}

	if strings.Contains(string(body), ">..<") {
		return result, err
	}
	storageObject := newHttpFileObject(URL, StorageObjectContentType, nil, &now, response.ContentLength)
	result = append(result, storageObject)
	return result, err
}


//Exists returns true if resource exists
func (s *httpStorageService) Exists(URL string) (bool, error) {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return false, err
	}
	if parsedUrl.Scheme != "file" {
		return false, fmt.Errorf("Invalid schema, expected file but had: %v", parsedUrl.Scheme)
	}
	return toolbox.FileExists(parsedUrl.Path), nil
}

//Object returns a Object for supplied url
func (s *httpStorageService) StorageObject(URL string) (Object, error) {
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
	return newFileObject(URL, objectType, &fileInfo, &modTime, fileInfo.Size()), nil
}

//Download returns reader for downloaded storage object
func (s *httpStorageService) Download(object Object) (io.Reader, error) {
	reader, _, err := toolbox.OpenReaderFromURL(s.addCredentialToURLIfNeeded(object.URL()))
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(content), err
}

//Upload uploads provided reader content for supplied url.
func (s *httpStorageService) Upload(URL string, reader io.Reader) error {
	return errors.New("unsupported")
}

func (s *httpStorageService) Register(schema string, service Service) error {
	return errors.New("unsupported")
}

//Delete removes passed in storage object
func (s *httpStorageService) Delete(object Object) error {
	fileName, err := toolbox.FileFromURL(object.URL())
	if err != nil {
		return err
	}
	return os.Remove(fileName)
}

func NewHttpStorageService(credential *PasswordCredential) Service {
	return &httpStorageService{
		Credential: credential,
	}
}

type httpStorageObject struct {
	*AbstractObject
}

func (o *httpStorageObject) Unwrap(target interface{}) error {
	return fmt.Errorf("unsuported target %T", target)
}

func newHttpFileObject(url string, objectType int, source interface{}, lastModified *time.Time, size int64) Object {
	abstract := NewAbstractStorageObject(url, source, objectType, lastModified, size)
	result := &httpStorageObject{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}

const HttpProviderScheme = "http"
const HttpsProviderScheme = "https"

func init() {
	NewStorageProvider().Registry[HttpProviderScheme] = serviceProvider
	NewStorageProvider().Registry[HttpsProviderScheme] = serviceProvider
}

func serviceProvider(credentialFile string) (Service, error) {

	if credentialFile == "" {
		return NewHttpStorageService(nil), nil
	}


	if !strings.HasPrefix(credentialFile, "/") {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		credentialFile = path.Join(dir, credentialFile)
	}
	config := &PasswordCredential{}
	err := toolbox.LoadConfigFromUrl("file://"+credentialFile, config)
	if err != nil {
		return nil, err
	}
	return NewHttpStorageService(config), nil
}
