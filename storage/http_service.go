package storage

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/cred"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//httpStorageService represents basic http storage service (only limited listing and full download are supported)
type httpStorageService struct {
	Credential *cred.Config
}

//HTTPClientProvider represents http client provider
var HTTPClientProvider = func() (*http.Client, error) {
	return toolbox.NewHttpClient(&toolbox.HttpOptions{Key: "MaxIdleConns", Value: 0})
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
		linkHref := string(linkContent[:linkEndPosition])
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
			URL:   linkHref,
			Value: strings.Trim(content, " \t\r\n"),
		}
		result = append(result, link)

	}
	return result
}

//List returns a list of object for supplied url
func (s *httpStorageService) List(URL string) ([]Object, error) {
	listURL := s.addCredentialToURLIfNeeded(URL)
	client, err := HTTPClientProvider()
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

	isGitUrl := strings.Contains(URL, "github.")
	isPublicGit := strings.Contains(URL, "github.com")
	if strings.Contains(contentType, "text/html") {

		links := extractLinks(string(body))
		var indexedLinks = map[string]bool{}
		if isGitUrl {

			for _, link := range links {
				if !((strings.Contains(link.URL, "/blob/") || strings.Contains(link.URL, "/tree/")) && strings.HasSuffix(link.URL, link.Value)) {
					continue
				}
				linkType := StorageObjectContentType
				_, name := toolbox.URLSplit(link.URL)
				if path.Ext(name) == "" {
					linkType = StorageObjectFolderType
				}

				baseURL := toolbox.URLBase(URL)

				objectURL := link.URL
				if !strings.Contains(objectURL, baseURL) {
					objectURL = toolbox.URLPathJoin(baseURL, link.URL)
				}

				if linkType == StorageObjectContentType && strings.Contains(objectURL, "/master/") {
					objectURL = strings.Replace(objectURL, "/blob/", "/", 1)
					if isPublicGit {
						objectURL = strings.Replace(objectURL, "github.com", "raw.githubusercontent.com", 1)
					} else {
						objectURL = strings.Replace(objectURL, ".com/", ".com/raw/", 1)
					}
				}
				if linkType == StorageObjectContentType && !strings.Contains(objectURL, "raw") {
					continue
				}
				if _, ok := indexedLinks[objectURL]; ok {
					continue
				}
				storageObject := newHttpFileObject(objectURL, linkType, nil, now, 1)
				indexedLinks[objectURL] = true
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
				storageObject := newHttpFileObject(objectURL, linkType, nil, now, 1)
				result = append(result, storageObject)
			}
		}
	}

	if strings.Contains(string(body), ">..<") {
		return result, err
	}
	storageObject := newHttpFileObject(URL, StorageObjectContentType, nil, now, response.ContentLength)
	result = append(result, storageObject)
	return result, err
}

//Exists returns true if resource exists
func (s *httpStorageService) Exists(URL string) (bool, error) {
	client, err := HTTPClientProvider()
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
		return nil, fmt.Errorf("resource  not found: %v", URL)
	}

	return objects[0], nil
}

//Download returns reader for downloaded storage object
func (s *httpStorageService) Download(object Object) (io.ReadCloser, error) {
	client, err := HTTPClientProvider()
	if err != nil {
		return nil, err
	}
	response, err := client.Get(s.addCredentialToURLIfNeeded(object.URL()))
	return response.Body, err
}

//Upload uploads provided reader content for supplied url.
func (s *httpStorageService) Upload(URL string, reader io.Reader) error {
	return errors.New("unsupported")
}

//Upload uploads provided reader content for supplied url.
func (s *httpStorageService) UploadWithMode(URL string, mode os.FileMode, reader io.Reader) error {
	return errors.New("unsupported")
}

func (s *httpStorageService) Register(schema string, service Service) error {
	return errors.New("unsupported")
}

//Delete removes passed in storage object
func (s *httpStorageService) Delete(object Object) error {
	fileName := toolbox.Filename(object.URL())
	return os.Remove(fileName)
}

func (s *httpStorageService) Close() error {
	return nil
}

//DownloadWithURL downloads content for passed in object URL
func (s *httpStorageService) DownloadWithURL(URL string) (io.ReadCloser, error) {
	object, err := s.StorageObject(URL)
	if err != nil {
		return nil, err
	}
	return s.Download(object)
}

func NewHttpStorageService(credential *cred.Config) Service {
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

func newHttpFileObject(url string, objectType int, source interface{}, lastModified time.Time, size int64) Object {
	var isDir = objectType == StorageObjectFolderType
	var _, name = toolbox.URLSplit(url)
	var fileMode, _ = NewFileMode("-r--r--r--")
	if isDir {
		fileMode, _ = NewFileMode("dr--r--r--")
	}
	fileInfo := NewFileInfo(name, size, fileMode, lastModified, isDir)
	abstract := NewAbstractStorageObject(url, source, fileInfo)
	result := &httpStorageObject{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}

const HttpProviderScheme = "http"
const HttpsProviderScheme = "https"

func init() {
	Registry().Registry[HttpsProviderScheme] = httpServiceProvider
	Registry().Registry[HttpProviderScheme] = httpServiceProvider

}

func httpServiceProvider(credentialFile string) (Service, error) {

	if credentialFile == "" {
		return NewHttpStorageService(nil), nil
	}

	if !strings.HasPrefix(credentialFile, "/") {
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		credentialFile = path.Join(dir, credentialFile)
	}
	config, err := cred.NewConfig(credentialFile)
	if err != nil {
		return nil, err
	}
	return NewHttpStorageService(config), nil
}
