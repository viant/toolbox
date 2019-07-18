package storage

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const MemoryProviderScheme = "mem"

var noSuchFileOrDirectoryError = errors.New("No such file or directory")
var folderMode, _ = NewFileMode("drwxrwxrwx")

//MemoryRoot represents memory root storage
var MemoryRoot = newMemoryFolder("mem:///", NewFileInfo("/", 102, folderMode, time.Now(), true))

//ResetMemory reset memory root storage
func ResetMemory() {
	MemoryRoot = newMemoryFolder("mem:///", NewFileInfo("/", 102, folderMode, time.Now(), true))
}

//Service represents memory storage service intended for testing
type memoryStorageService struct {
	root *MemoryFolder
}

type MemoryFile struct {
	name     string
	fileInfo os.FileInfo
	content  []byte
}

func (f *MemoryFile) Object() Object {
	return NewAbstractStorageObject(f.name, f, f.fileInfo)
}

type MemoryFolder struct {
	name     string
	fileInfo os.FileInfo
	mutext   *sync.RWMutex
	files    map[string]*MemoryFile
	folders  map[string]*MemoryFolder
}

func (f *MemoryFolder) Object() Object {
	return NewAbstractStorageObject(f.name, f, f.fileInfo)
}

func (f *MemoryFolder) Objects() []Object {
	var result = make([]Object, 0)
	result = append(result, f.Object())
	for _, folder := range f.folders {
		result = append(result, folder.Object())
	}
	for _, file := range f.files {
		result = append(result, file.Object())
	}
	return result
}

func newMemoryFolder(name string, info os.FileInfo) *MemoryFolder {
	return &MemoryFolder{
		name:     name,
		fileInfo: info,
		mutext:   &sync.RWMutex{},
		files:    make(map[string]*MemoryFile),
		folders:  make(map[string]*MemoryFolder),
	}
}

func (s *memoryStorageService) getFolder(pathFragments []string) (*MemoryFolder, error) {
	node := s.root
	var ok bool
	for i := 1; i+1 < len(pathFragments); i++ {
		pathFragment := pathFragments[i]
		node, ok = node.folders[pathFragment]
		if !ok {
			return nil, noSuchFileOrDirectoryError
		}
	}
	return node, nil
}

func (s *memoryStorageService) getPath(URL string) (string, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return "", err
	}
	parsedPath := parsedURL.Path
	strings.Replace(parsedPath, "//", "/", len(parsedPath))
	if len(parsedPath) > 1 && strings.HasSuffix(parsedPath, "/") {
		parsedPath = string(parsedPath[:len(parsedPath)-1])
	}
	return parsedPath, nil
}

//List returns a list of object for supplied url
func (s *memoryStorageService) List(URL string) ([]Object, error) {
	path, err := s.getPath(URL)
	if err != nil {
		return nil, err
	}
	if path == "/" {
		return s.root.Objects(), nil
	}
	var pathFragments = strings.Split(path, "/")
	node, err := s.getFolder(pathFragments)
	if err != nil {
		return nil, err
	}
	var pathLeaf = pathFragments[len(pathFragments)-1]

	if memoryFile, ok := node.files[pathLeaf]; ok {
		return []Object{memoryFile.Object()}, nil
	}
	if folder, ok := node.folders[pathLeaf]; ok {
		return folder.Objects(), nil
	}

	return []Object{}, nil
}

//Exists returns true if resource exists
func (s *memoryStorageService) Exists(URL string) (bool, error) {
	objects, err := s.List(URL)
	if err != nil {
		return false, err
	}
	return len(objects) > 0, nil
}

func (s *memoryStorageService) Close() error {
	return nil
}

//Object returns a Object for supplied url
func (s *memoryStorageService) StorageObject(URL string) (Object, error) {
	objects, err := s.List(URL)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, noSuchFileOrDirectoryError
	}
	return objects[0], nil
}

//Download returns reader for downloaded storage object
func (s *memoryStorageService) Download(object Object) (io.ReadCloser, error) {
	var urlPath, err = s.getPath(object.URL())
	if err != nil {
		return nil, err
	}
	var pathFragments = strings.Split(urlPath, "/")
	node, err := s.getFolder(pathFragments)
	if err != nil {
		return nil, err
	}
	var pathLeaf = pathFragments[len(pathFragments)-1]
	if memoryFile, ok := node.files[pathLeaf]; ok {
		return ioutil.NopCloser(bytes.NewReader(memoryFile.content)), nil
	}
	return nil, noSuchFileOrDirectoryError
}

//Upload uploads provided reader content for supplied url.
func (s *memoryStorageService) Upload(URL string, reader io.Reader) error {
	return s.UploadWithMode(URL, DefaultFileMode, reader)
}

//Upload uploads provided reader content for supplied url.
func (s *memoryStorageService) UploadWithMode(URL string, mode os.FileMode, reader io.Reader) error {
	urlPath, err := s.getPath(URL)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var node = s.root
	var pathFragments = strings.Split(urlPath, "/")
	for i := 1; i+1 < len(pathFragments); i++ {
		pathFragment := pathFragments[i]
		if subFolder, ok := node.folders[pathFragment]; ok {
			node = subFolder
		} else {
			var folderURL = MemoryProviderScheme + "://" + strings.Join(pathFragments[:i+1], "/")
			var folderInfo = NewFileInfo(pathFragment, 102, folderMode, time.Now(), true)
			newFolder := newMemoryFolder(folderURL, folderInfo)
			node.mutext.Lock()
			node.folders[folderInfo.Name()] = newFolder
			node.mutext.Unlock()
			node = newFolder
		}
	}

	var pathLeaf = pathFragments[len(pathFragments)-1]
	fileInfo := NewFileInfo(pathLeaf, int64(len(content)), fileMode, time.Now(), false)
	var memoryFile = &MemoryFile{name: URL, content: content, fileInfo: fileInfo}
	node.files[fileInfo.Name()] = memoryFile
	return nil
}

func (s *memoryStorageService) Register(schema string, service Service) error {
	return errors.New("unsupported")
}

//Delete removes passed in storage object
func (s *memoryStorageService) Delete(object Object) error {
	var urlPath, err = s.getPath(object.URL())
	if err != nil {
		return err
	}
	var pathFragments = strings.Split(urlPath, "/")
	node, err := s.getFolder(pathFragments)
	if err != nil {
		return err
	}
	var pathLeaf = pathFragments[len(pathFragments)-1]
	if _, ok := node.files[pathLeaf]; ok {
		delete(node.files, pathLeaf)
		return nil
	}
	if _, ok := node.files[pathLeaf]; ok {
		delete(node.folders, pathLeaf)
		return nil
	}
	return noSuchFileOrDirectoryError
}

//DownloadWithURL downloads content for passed in object URL
func (s *memoryStorageService) DownloadWithURL(URL string) (io.ReadCloser, error) {
	object, err := s.StorageObject(URL)
	if err != nil {
		return nil, err
	}
	return s.Download(object)
}

// creates a new private memory service
func NewPrivateMemoryService() Service {
	return &memoryStorageService{
		root: newMemoryFolder("mem:///", NewFileInfo("/", 102, folderMode, time.Now(), true)),
	}
}

//creates a new memory service
func NewMemoryService() Service {
	return &memoryStorageService{
		root: MemoryRoot,
	}
}

func init() {
	Registry().Registry[MemoryProviderScheme] = memServiceProvider
}

func memServiceProvider(credentialFile string) (Service, error) {
	return NewMemoryService(), nil
}
