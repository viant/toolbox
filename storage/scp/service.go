package scp

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/lunixbochs/vtclean"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/cred"
	"github.com/viant/toolbox/ssh"
	"github.com/viant/toolbox/storage"
	"io"
	"io/ioutil"
	"net/url"
	"path"
	"strings"
	"sync"
)

const defaultSSHPort = 22
const verificationSizeThreshold = 1024 * 1024

//NoSuchFileOrDirectoryError represents no such file or directory error
var NoSuchFileOrDirectoryError = errors.New("No Such File Or Directory")

type service struct {
	fileService   storage.Service
	config        *cred.Config
	services      map[string]ssh.Service
	multiSessions map[string]ssh.MultiCommandSession
	mutex         *sync.Mutex
}

func (s *service) runCommand(session ssh.MultiCommandSession, URL string, command string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	output, _ := session.Run(command, 5000)
	var stdout = s.stdout(output)
	return stdout, nil
}

func (s *service) stdout(output string) string {
	return vtclean.Clean(string(output), false)
}

func (s *service) getMultiSession(parsedURL *url.URL) ssh.MultiCommandSession {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.multiSessions[parsedURL.Host]
}

func (s *service) getService(parsedURL *url.URL) (ssh.Service, error) {
	port := toolbox.AsInt(parsedURL.Port())
	if port == 0 {
		port = 22
	}
	key := parsedURL.Host
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if service, ok := s.services[key]; ok {
		return service, nil
	}
	service, err := ssh.NewService(parsedURL.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return nil, err
	}
	s.services[key] = service
	s.multiSessions[key], err = service.OpenMultiCommandSession(nil)
	if err != nil {
		return nil, err
	}
	return service, nil
}

//List returns a list of object for supplied URL
func (s *service) List(URL string) ([]storage.Object, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Host == "127.0.0.1" || parsedURL.Host == "127.0.0.1:22" {
		var fileURL = toolbox.FileSchema + parsedURL.Path
		return s.fileService.List(fileURL)
	}
	_, err = s.getService(parsedURL)
	if err != nil {
		return nil, err
	}
	commandSession := s.getMultiSession(parsedURL)
	canListWithTimeStyle := commandSession.System() != "darwin"
	var parser = &Parser{IsoTimeStyle: canListWithTimeStyle}
	var URLPath = parsedURL.Path
	var result = make([]storage.Object, 0)
	var lsCommand = "ls -dltr"
	if canListWithTimeStyle {
		lsCommand += " --time-style=full-iso"
	} else {
		lsCommand += "T"
	}
	output, _ := s.runCommand(commandSession, URL, lsCommand+" "+URLPath)
	var stdout = vtclean.Clean(string(output), false)
	if strings.Contains(stdout, "No such file or directory") {
		return result, NoSuchFileOrDirectoryError
	}
	objects, err := parser.Parse(parsedURL, stdout, false)
	if err != nil {
		return nil, err
	}
	if len(objects) == 1 && objects[0].FileInfo().IsDir() {
		output, _ = s.runCommand(commandSession, URL, lsCommand+" "+path.Join(URLPath, "*"))
		stdout = vtclean.Clean(string(output), false)
		directoryObjects, err := parser.Parse(parsedURL, stdout, true)
		if err != nil {
			return nil, err
		}
		if len(directoryObjects) > 0 {
			objects = append(objects, directoryObjects...)
		}
	}
	return objects, nil
}

func (s *service) Exists(URL string) (bool, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return false, err
	}

	if parsedURL.Host == "127.0.0.1" || parsedURL.Host == "127.0.0.1:22" {
		var fileURL = toolbox.FileSchema + parsedURL.Path
		return s.fileService.Exists(fileURL)
	}
	_, err = s.getService(parsedURL)
	if err != nil {
		return false, err
	}
	commandSession := s.getMultiSession(parsedURL)
	output, _ := s.runCommand(commandSession, URL, "ls -dltr "+parsedURL.Path)
	if strings.Contains(string(output), "No such file or directory") {
		return false, nil
	}
	return true, nil

}

func (s *service) StorageObject(URL string) (storage.Object, error) {
	objects, err := s.List(URL)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, NoSuchFileOrDirectoryError
	}
	return objects[0], nil
}

//Download returns reader for downloaded storage object
func (s *service) Download(object storage.Object) (io.ReadCloser, error) {
	if object == nil {
		return nil, fmt.Errorf("Object was nil")
	}
	parsedURL, err := url.Parse(object.URL())
	if err != nil {
		return nil, err
	}
	if parsedURL.Host == "127.0.0.1" || parsedURL.Host == "127.0.0.1:22" {
		var fileURL = toolbox.FileSchema + parsedURL.Path
		storageObject, err := s.fileService.StorageObject(fileURL)
		if err != nil {
			return nil, err
		}
		return s.fileService.Download(storageObject)
	}
	port := toolbox.AsInt(parsedURL.Port())
	if port == 0 {
		port = defaultSSHPort
	}

	service, err := s.getService(parsedURL)
	if err != nil {
		return nil, err
	}
	content, err := service.Download(parsedURL.Path)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(content)), nil
}

//Upload uploads provided reader content for supplied URL.
func (s *service) Upload(URL string, reader io.Reader) error {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if parsedURL.Host == "127.0.0.1" || parsedURL.Host == "127.0.0.1:22" {
		var fileURL = toolbox.FileSchema + parsedURL.Path
		return s.fileService.Upload(fileURL, reader)
	}

	port := toolbox.AsInt(parsedURL.Port())
	if port == 0 {
		port = defaultSSHPort
	}
	service, err := ssh.NewService(parsedURL.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return err
	}

	//defer service.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to upload - unable read: %v", err)
	}

	err = service.Upload(parsedURL.Path, content)
	if err != nil {
		return fmt.Errorf("failed to upload: %v %v", URL, err)
	}

	if verificationSizeThreshold < len(content) {
		object, err := s.StorageObject(URL)
		if err != nil {
			return fmt.Errorf("failed to get upload object  %v for verification: %v", URL, err)
		}
		if int(object.FileInfo().Size()) != len(content) {
			err = service.Upload(parsedURL.Path, content)
			object, err = s.StorageObject(URL)
			if err != nil {
				return err
			}
			if int(object.FileInfo().Size()) != len(content) {
				return fmt.Errorf("failed to upload to %v, actual size was:%v,  but uploaded size was %v", URL, len(content), int(object.FileInfo().Size()))
			}
		}
	}

	return err
}

func (s *service) Register(schema string, service storage.Service) error {
	return errors.New("unsupported")
}

func (s *service) Close() error {
	for _, service := range s.services {
		service.Close()
	}
	for _, session := range s.multiSessions {
		session.Close()
	}
	return nil
}

//Delete removes passed in storage object
func (s *service) Delete(object storage.Object) error {
	parsedURL, err := url.Parse(object.URL())
	if err != nil {
		return err
	}
	if parsedURL.Host == "127.0.0.1" || parsedURL.Host == "127.0.0.1:22" {
		var fileURL = toolbox.FileSchema + parsedURL.Path
		storageObject, err := s.fileService.StorageObject(fileURL)
		if err != nil {
			return err
		}
		return s.fileService.Delete(storageObject)
	}

	port := toolbox.AsInt(parsedURL.Port())
	if port == 0 {
		port = defaultSSHPort
	}
	service, err := ssh.NewService(parsedURL.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return err
	}

	//defer service.Close()
	session, err := service.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if parsedURL.Path == "/" {
		return fmt.Errorf("Invalid removal path: %v", parsedURL.Path)
	}
	_, err = session.Output("rm -rf " + parsedURL.Path)
	return err
}

//NewService create a new gc storage service
func NewService(config *cred.Config) *service {
	return &service{
		services:      make(map[string]ssh.Service),
		config:        config,
		multiSessions: make(map[string]ssh.MultiCommandSession),
		mutex:         &sync.Mutex{},
		fileService:   storage.NewFileStorage(),
	}
}
