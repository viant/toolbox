package scp

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/ssh"
	"github.com/viant/toolbox/storage"
	"io"
	"io/ioutil"
	"strings"
	"github.com/viant/toolbox/cred"
	"net/url"
	"github.com/lunixbochs/vtclean"
	"path"
	"sync"
)

const defaultSSHPort = 22
const verificationSizeThreshold = 1024 * 1024

//NoSuchFileOrDirectoryError represents no such file or directory error
var NoSuchFileOrDirectoryError = errors.New("No Such File Or Directory")

type service struct {
	config   *cred.Config
	services map[string]ssh.Service
	multiSessions map[string]ssh.MultiCommandSession
	mutex    *sync.Mutex
}

func (s *service) runCommand(session ssh.MultiCommandSession, URL string, command string) (string, error) {
	output, _ := session.Run(command, 1000)
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
func (s *service) Download(object storage.Object) (io.Reader, error) {
	if object == nil {
		return nil, fmt.Errorf("Object was nil")
	}
	parsedUrl, err := url.Parse(object.URL())
	if err != nil {
		return nil, err
	}

	port := toolbox.AsInt(parsedUrl.Port())
	if port == 0 {
		port = defaultSSHPort
	}

	service, err := s.getService(parsedUrl)
	content, err := service.Download(parsedUrl.Path)
	if err != nil {
		return nil, err
	}



	if  verificationSizeThreshold < len(content) {
		//download verification (as sometimes scp failed) with one retry
		if int(object.FileInfo().Size()) != len(content) {
			content, err = service.Download(parsedUrl.Path)
			if err != nil {
				return nil, err
			}
			if int(object.FileInfo().Size()) != len(content) {
				return nil, fmt.Errorf("Faled to download from %v,  object size was: %v, but scp download was %v", object.URL(), object.FileInfo().Size(), len(content))
			}
		}
	}



	return bytes.NewReader(content), nil
}

//Upload uploads provided reader content for supplied URL.
func (s *service) Upload(URL string, reader io.Reader) error {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return err
	}

	port := toolbox.AsInt(parsedUrl.Port())
	if port == 0 {
		port = defaultSSHPort
	}
	service, err := ssh.NewService(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return err
	}

	//defer service.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Failed to upload - unable read: %v", err)
	}

	err = service.Upload(parsedUrl.Path, content)
	if err != nil {
		return fmt.Errorf("Failed to upload: %v %v", URL, err)
	}

	if verificationSizeThreshold < len(content) {
		object, err := s.StorageObject(URL)
		if err != nil {
			return fmt.Errorf("Failed to get upload object  %v for verification: %v", URL, err)
		}
		if int(object.FileInfo().Size()) != len(content) {
			err = service.Upload(parsedUrl.Path, content)
			object, err = s.StorageObject(URL)
			if err != nil {
				return err
			}
			if int(object.FileInfo().Size()) != len(content) {
				return fmt.Errorf("Failed to upload to %v, actual size was:%v,  but uploaded size was %v", URL, len(content), int(object.FileInfo().Size()))
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
	parsedUrl, err := url.Parse(object.URL())
	if err != nil {
		return err
	}

	port := toolbox.AsInt(parsedUrl.Port())
	if port == 0 {
		port = defaultSSHPort
	}
	service, err := ssh.NewService(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return err
	}

	//defer service.Close()
	session, err := service.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if parsedUrl.Path == "/" {
		return fmt.Errorf("Invalid removal path: %v", parsedUrl.Path)
	}
	_, err = session.Output("rm -rf " + parsedUrl.Path)
	return err
}

//NewService create a new gc storage service
func NewService(config *cred.Config) *service {
	return &service{
		services: make(map[string]ssh.Service),
		config:   config,
		multiSessions: make(map[string]ssh.MultiCommandSession),
		mutex:    &sync.Mutex{},
	}
}
