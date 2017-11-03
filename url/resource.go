package url

import (
	"bytes"
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/storage"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

//Resource represents a URL based resource, with enriched meta info
type Resource struct {
	URL           string   //URL of resource
	Credential    string   //name of file or alias to the file defined via credential service
	ParsedURL     *url.URL //parsed URL resource
	Cache         string   //Cache path for the resource, if specified resource will be cached in the specified path
	CacheExpiryMs int      //CacheExpiryMs expiry time in ms

	Name    string //name of a resource
	Version string //version of resource
	Type    string //resource type
}

//Clone creates a clone of the resource
func (r *Resource) Clone() *Resource {
	return &Resource{
		Name:          r.Name,
		Version:       r.Version,
		URL:           r.URL,
		Type:          r.Type,
		Credential:    r.Credential,
		ParsedURL:     r.ParsedURL,
		Cache:         r.Cache,
		CacheExpiryMs: r.CacheExpiryMs,
	}
}

var defaultSchemePorts = map[string]int{
	"ssh":   22,
	"http":  80,
	"https": 443,
}

//Host returns url's host name with user name if user name is part of url
func (r *Resource) Host() string {
	result := r.ParsedURL.Hostname() + ":" + r.Port()
	if r.ParsedURL.User != nil {
		result = r.ParsedURL.User.Username() + "@" + result
	}
	return result
}

//Port returns url's port
func (r *Resource) Port() string {
	port := r.ParsedURL.Port()
	if port == "" && r.ParsedURL != nil {
		if value, ok := defaultSchemePorts[r.ParsedURL.Scheme]; ok {
			port = toolbox.AsString(value)
		}
	}
	return port
}

//LoadCredentialload credential, returns username, password. It takes errorIfEmpty flag to return an error if there is issue with credential
func (r *Resource) LoadCredential(errorIfEmpty bool) (string, string, error) {
	if r.Credential == "" {
		if errorIfEmpty {
			return "", "", fmt.Errorf("Credential was empty: %v", r.Credential)
		}
		return "", "", nil
	}
	credential, err := cred.NewConfig(r.Credential)
	if err != nil {
		return "", "", fmt.Errorf("Failed to load credential: %v %v", r.Credential, err)
	}
	return credential.Username, credential.Password, nil
}

//Download downloads data from URL, it returns data as []byte, or error, if resource is cacheable it first look into cache
func (r *Resource) Download() ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("Fail to download content on empty resource")
	}

	if r.Cachable() {
		content := r.readFromCache()
		if content != nil {
			return content, nil
		}
	}

	service, err := storage.NewServiceForURL(r.URL, r.Credential)
	if err != nil {
		return nil, err
	}
	object, err := service.StorageObject(r.URL)
	if err != nil {
		return nil, err
	}
	reader, err := service.Download(object)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if r.Cachable() {
		_ = ioutil.WriteFile(r.Cache, content, 0666)
	}
	return content, err
}

//DownloadText returns a text downloaded from url
func (r *Resource) DownloadText() (string, error) {
	var result, err = r.Download()
	if err != nil {
		return "", err
	}
	return string(result), err
}

//Decode decodes url's data into target, it takes decoderFactory which decodes data into target
func (r *Resource) Decode(target interface{}, decoderFactory toolbox.DecoderFactory) error {
	if r == nil {
		return fmt.Errorf("Fail to %T decode on empty resource", decoderFactory)
	}
	if r == nil {
		return fmt.Errorf("Fail to decode %v,  decoderFactory was empty", r.URL, decoderFactory)
	}
	var content, err = r.Download()
	if err != nil {
		return err
	}
	return decoderFactory.Create(bytes.NewReader(content)).Decode(target)
}

//JSONDecode decodes json resource into target
func (r *Resource) JSONDecode(target interface{}) error {
	return r.Decode(target, toolbox.NewJSONDecoderFactory())
}

//JSONDecode decodes yaml resource into target
func (r *Resource) YAMLDecode(target interface{}) error {
	return r.Decode(target, toolbox.NewYamlDecoderFactory())
}

func (r *Resource) readFromCache() []byte {
	if toolbox.FileExists(r.Cache) {
		info, err := os.Stat(r.Cache)
		var isExpired = false
		if err == nil && r.CacheExpiryMs > 0 {
			elapsed := time.Now().Sub(info.ModTime())
			isExpired = elapsed > time.Second*time.Duration(r.CacheExpiryMs)
		}
		content, err := ioutil.ReadFile(r.Cache)
		if err == nil && !isExpired {
			return content
		}
	}
	return nil
}

//Cachable returns true if resource is cachable
func (r *Resource) Cachable() bool {
	return r.Cache != ""
}

func normalizeURL(URL string) string {
	if strings.Contains(URL, "://") {
		return URL
	}
	if !strings.HasPrefix(URL, "/") {
		currentDirectory, err := os.Getwd()
		if err == nil {
			candidate := path.Join(currentDirectory, URL)
			if toolbox.FileExists(candidate) {
				URL = candidate
			}
		}
	}
	return toolbox.FileSchema + URL
}

//NewResource returns a new resource for provided URL, followed by optional credential, cache and cache expiryMs.
func NewResource(Params ...interface{}) *Resource {
	if len(Params) == 0 {
		return nil
	}
	var URL = toolbox.AsString(Params[0])
	URL = normalizeURL(URL)

	var credential string
	if len(Params) > 1 {
		credential = toolbox.AsString(Params[1])
	}
	var cache string
	if len(Params) > 2 {
		cache = toolbox.AsString(Params[2])
	}
	var cacheExpiryMs int
	if len(Params) > 3 {
		cacheExpiryMs = toolbox.AsInt(Params[3])
	}
	parsedURL, _ := url.Parse(URL)
	return &Resource{
		ParsedURL:     parsedURL,
		URL:           URL,
		Credential:    credential,
		Cache:         cache,
		CacheExpiryMs: cacheExpiryMs,
	}
}
