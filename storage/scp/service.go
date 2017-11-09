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
)

const defaultSSHPort = 22
const verificationSizeThreshold = 1024 * 1024

//NoSuchFileOrDirectoryError represents no such file or directory error
var NoSuchFileOrDirectoryError = errors.New("No Such File Or Directory")

type service struct {
	config *cred.Config
}

func (s *service) runCommand(session ssh.MultiCommandSession, URL string, command string) (string, error) {
	output, _ := session.Run(command, 0, "$ ", "usage", "No such file or directory")
	return toolbox.AsString(output), nil
}

func (s *service) canListWithTimeStyle(session ssh.MultiCommandSession, URL string) (bool) {
	return session.KernelName() != "darwin"
}

func (s *service) getClient(parsedURL *url.URL) (ssh.Service, error) {
	port := toolbox.AsInt(parsedURL.Port())
	if port == 0 {
		port = 22
	}
	return ssh.NewService(parsedURL.Hostname(), toolbox.AsInt(port), s.config)
}

//List returns a list of object for supplied URL
func (s *service) List(URL string) ([]storage.Object, error) {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	client, err := s.getClient(parsedUrl)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	session, err := client.OpenMultiCommandSession(&ssh.SessionConfig{})
	if err != nil {
		return nil, err
	}
	defer session.Close()
	canListWithTimeStyle := s.canListWithTimeStyle(session, URL)
	var parser = &Parser{IsoTimeStyle: canListWithTimeStyle}
	var urlPath = strings.Replace(parsedUrl.Path, "//", "/", len(parsedUrl.Path))
	var result = make([]storage.Object, 0)
	var lsCommand = "ls -dltr"
	if canListWithTimeStyle {
		lsCommand += " --time-style=full-iso"
	} else {
		lsCommand += "T"
	}
	output, err := s.runCommand(session, URL, lsCommand+" "+urlPath)
	var stdout = vtclean.Clean(string(output), false)
	if strings.Contains(stdout, "No such file or directory") {
		return result, NoSuchFileOrDirectoryError
	}
	objects, err := parser.Parse(URL, stdout, false)
	if err != nil {
		return nil, err
	}
	if len(objects) == 1 && objects[0].FileInfo().IsDir() {
		output, err = s.runCommand(session, URL, lsCommand+" "+urlPath+"/*")
		stdout = vtclean.Clean(string(output), false)
		directoryObjects, err := parser.Parse(URL, stdout, true)
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
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return false, err
	}

	client, err := s.getClient(parsedUrl)
	if err != nil {
		return false, err
	}
	defer client.Close()
	session, err := client.OpenMultiCommandSession(&ssh.SessionConfig{})
	if err != nil {
		return false, err
	}
	defer session.Close()

	output, err := s.runCommand(session, URL, "ls -dltr "+parsedUrl.Path)
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
	client, err := ssh.NewService(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	content, err := client.Download(parsedUrl.Path)
	if err != nil {
		return nil, err
	}

	if verificationSizeThreshold < len(content) {
		//download verification (as sometimes scp failed) with one retry
		if int(object.FileInfo().Size()) != len(content) {
			content, err = client.Download(parsedUrl.Path)
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
	client, err := ssh.NewService(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return err
	}
	defer client.Close()

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Failed to upload - unable read: %v", err)
	}

	err = client.Upload(parsedUrl.Path, content)
	if err != nil {
		return fmt.Errorf("Failed to upload: %v %v", URL, err)
	}

	if verificationSizeThreshold < len(content) {
		object, err := s.StorageObject(URL)
		if err != nil {
			return fmt.Errorf("Failed to get upload object  %v for verification: %v", URL, err)
		}
		if int(object.FileInfo().Size()) != len(content) {
			err = client.Upload(parsedUrl.Path, content)
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
	client, err := ssh.NewService(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return err
	}
	defer client.Close()
	session, err := client.NewSession()
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
		config: config,
	}
}
