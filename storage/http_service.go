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


func newHttpClient() (*http.Client, error) {
	return toolbox.NewHttpClient(&toolbox.HttpOptions{Key:"MaxIdleConns", Value:0})
}

func (s *httpStorageService) addCredentialToURLIfNeeded(URL string) string {
	if s.Credential == nil || s.Credential.Password == "" || s.Credential.Username == "" {
		return URL
	}
	prasedURL, err := url.Parse(URL)
	if err != nil {
		return URL
	}
	if prasedURL.User != nil {
		return URL
	}
	return strings.Replace(URL, "://", fmt.Sprintf("://%v:%v@", s.Credential.Username, s.Credential.Password), 1)
}

type hRef struct {
	URL   string
	Value string
}

func extractLinks(body string) []*hRef {
	var result = make([]*hRef, 0)
	var linkContents = strings.Split(string(body), "href=\"")
	for i := 1; i < len(linkContents); i++ {
		var linkContent = linkContents[i]
		linkEndPosition := strings.Index(linkContent, "\"")
		if linkEndPosition == -1 {
			continue
		}
		var content = ""
		contentStartPosition := strings.Index(linkContent, ">")
		if contentStartPosition != 1 {
			content = string(linkContent[contentStartPosition+1:])
			contentEndPosition := strings.Index(content, "<")
			if contentEndPosition != -1 {
				content = string(content[:contentEndPosition])
			}
		}

		link := &hRef{
			URL:   string(linkContent[:linkEndPosition]),
			Value: strings.Trim(content, " \t\r\n"),
		}
		result = append(result, link)

	}
	return result
}

//List returns a list of object for supplied url
func (s *httpStorageService) List(URL string) ([]Object, error) {
	listURL := s.addCredentialToURLIfNeeded(URL)
	client, err := newHttpClient()
	if err != nil {
		return nil, err
	}
	response, err := client.Get(listURL)

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

	isGitUrl := strings.Contains(URL, "github.com")
	if strings.Contains(contentType, "text/html") {

		links := extractLinks(string(body))

		if isGitUrl {

			for _, link := range links {
				if ! ((strings.Contains(link.URL, "/blob/") || strings.Contains(link.URL, "/tree/")) && strings.HasSuffix(link.URL, link.Value)) {
					continue
				}

				linkType := StorageObjectContentType
				if strings.Contains(link.URL, "/tree/") {
					linkType = StorageObjectFolderType
				}

				baseURL := toolbox.URLBase(URL)
				objectURL := toolbox.URLPathJoin(baseURL, link.URL)

				if linkType == StorageObjectContentType {
					objectURL = strings.Replace(objectURL, "/blob/", "/", 1)
					objectURL = strings.Replace(objectURL, "github.com", "raw.githubusercontent.com", 1)
				}
				storageObject := newHttpFileObject(objectURL, linkType, nil, &now, 0)
				result = append(result, storageObject)
			}

		} else {
			for _, link := range links {
				if link.URL == "" || strings.Contains(link.URL, ":") || strings.HasPrefix(link.URL, "#") || strings.HasPrefix(link.URL, "?") || strings.HasPrefix(link.URL, ".") || strings.HasPrefix(link.URL, "/") {
					continue
				}
				linkType := StorageObjectContentType
				if strings.HasSuffix(link.URL, "/") {
					linkType = StorageObjectFolderType
				}
				objectURL := toolbox.URLPathJoin(URL, link.URL)
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
	client, err := newHttpClient()
	if err != nil {
		return false, err
	}
	response, err := client.Get(URL)
	if err != nil {
		return false, err
	}
	return response.StatusCode == 200, nil
}


//Object returns a Object for supplied url
func (s *httpStorageService) StorageObject(URL string) (Object, error) {
	objects, err := s.List(URL)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("Resource not found: %v", URL)
	}

	return objects[0], nil
}

//Download returns reader for downloaded storage object
func (s *httpStorageService) Download(object Object) (io.Reader, error) {
	client, err := newHttpClient()
	if err != nil {
		return nil, err
	}
	response, err := client.Get(s.addCredentialToURLIfNeeded(object.URL()))
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
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
	NewStorageProvider().Registry[HttpsProviderScheme] = serviceProvider
	NewStorageProvider().Registry[HttpProviderScheme] = serviceProvider

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
