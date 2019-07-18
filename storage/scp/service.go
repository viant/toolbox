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
	"os"
	"path"
	"strings"
	"sync"
)

const (
	defaultSSHPort            = 22
	verificationSizeThreshold = 1024 * 1024
)

//NoSuchFileOrDirectoryError represents no such file or directory error
var NoSuchFileOrDirectoryError = errors.New("No such file or directory")

const unrecognizedOption = "unrecognized option"

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
	output, _ := session.Run(command, nil, 5000)
	var stdout = s.stdout(output)
	return stdout, nil
}

func (s *service) stdout(output string) string {
	var result = make([]string, 0)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		result = append(result, vtclean.Clean(line, false))
	}
	return strings.Join(result, "\n")
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
	var lsCommand = ""

	if canListWithTimeStyle {
		lsCommand += "ls -dltr --time-style=full-iso " + URLPath
	} else {
		lsCommand += "ls -dltrT " + URLPath
	}
	output, _ := s.runCommand(commandSession, URL, lsCommand)
	var stdout = vtclean.Clean(string(output), false)
	if strings.Contains(stdout, "unrecognized option") {
		if canListWithTimeStyle {
			lsCommand = "ls -dltr --full-time " + URLPath
			output, _ = s.runCommand(commandSession, URL, lsCommand)
			stdout = vtclean.Clean(string(output), false)
		}
	}

	if strings.Contains(stdout, unrecognizedOption) {
		return nil, fmt.Errorf("unable to list files with: %v, %v", lsCommand, stdout)
	}

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
		return nil, fmt.Errorf("object was nil")
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
	return s.UploadWithMode(URL, storage.DefaultFileMode, reader)
}

//Upload uploads provided reader content for supplied URL.
func (s *service) UploadWithMode(URL string, mode os.FileMode, reader io.Reader) error {
	if mode == 0 {
		mode = storage.DefaultFileMode
	}
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return err
	}

	if parsedURL.Host == "127.0.0.1" || parsedURL.Host == "127.0.0.1:22" {
		var fileURL = toolbox.FileSchema + parsedURL.Path
		return s.fileService.UploadWithMode(fileURL, mode, reader)
	}

	port := toolbox.AsInt(parsedURL.Port())
	if port == 0 {
		port = defaultSSHPort
	}

	//service, err := ssh.NewService(parsedURL.Hostname(), toolbox.AsInt(port), s.config)
	service, err := s.getService(parsedURL)
	if err != nil {
		return err
	}

	//defer service.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to upload - unable read: %v", err)
	}

	err = service.Upload(parsedURL.Path, mode, content)
	if err != nil {
		return fmt.Errorf("failed to upload: %v %v", URL, err)
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
		return fmt.Errorf("invalid removal path: %v", parsedURL.Path)
	}
	_, err = session.Output("rm -rf " + parsedURL.Path)
	return err
}

//DownloadWithURL downloads content for passed in object URL
func (s *service) DownloadWithURL(URL string) (io.ReadCloser, error) {
	object, err := s.StorageObject(URL)
	if err != nil {
		return nil, err
	}
	return s.Download(object)
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
