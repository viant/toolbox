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
	"net/url"
	"strings"
)

const defaultSSHPort = 22

const (
	fileInfoPermission = iota
	_
	fileInfoOwner
	fileInfoGroup
	fileInfoSize
	fileInfoDateMonth
	fileInfoDateDay
	fileInfoDateHour
	fileInfoDateYear
	fileInfoName
)

type service struct {
	config *ssh.AuthConfig
}

//List returns a list of object for supplied URL
func (s *service) List(URL string) ([]storage.Object, error) {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}


	port := toolbox.AsInt(parsedUrl.Port())
	if port == 0 {
		port = 22
	}
	client, err := ssh.NewClient(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var result = make([]storage.Object, 0)

	output, err := session.CombinedOutput("ls -lTtr " + parsedUrl.Path)
	if strings.Contains(string(output), "No such file or directory") {
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(output), "\n") {
		fileInfo := extractFileInfo(line)
		if fileInfo.name == "" {
			continue
		}
		fileInfo.url = URL
		result = append(result, fileInfo)
	}
	return result, nil
}

func extractFileInfo(line string) *object {
	fragmentCount := 0
	fileInfo := &object{}
	for i := range line {

		aChar := string(line[i])
		if aChar == " " || aChar == "\t" {
			if i+1 < len(line) {
				nextChar := string(line[i+1])
				if !(nextChar == " " || nextChar == "\t") {
					fragmentCount++
				}
			}
			continue
		}
		switch fragmentCount {

		case fileInfoPermission:
			fileInfo.permission += aChar
		case fileInfoOwner:
			fileInfo.owner += aChar
		case fileInfoGroup:
			fileInfo.group += aChar
		case fileInfoSize:
			fileInfo.size += aChar
		case fileInfoDateMonth:
			fileInfo.month += aChar
		case fileInfoDateDay:
			fileInfo.day += aChar
		case fileInfoDateHour:
			fileInfo.hour += aChar
		case fileInfoDateYear:
			fileInfo.year += aChar
		case fileInfoName:
			fileInfo.name += aChar
		}

	}
	return fileInfo
}

func (s *service) Exists(URL string) (bool, error) {
	objects, err := s.List(URL)
	if err != nil {
		return false, err
	}
	return len(objects) > 0, nil
}

func (s *service) StorageObject(URL string) (storage.Object, error) {
	objects, err := s.List(URL)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("No found %v", URL)
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
	client, err := ssh.NewClient(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	content, err := client.Download(parsedUrl.Path)
	if err != nil {
		return nil, err
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
	client, err := ssh.NewClient(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
	if err != nil {
		return err
	}
	defer client.Close()

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Failed to upload - unable read: %v", err)
	}
	return client.Upload(parsedUrl.Path, content)
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
	client, err := ssh.NewClient(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
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
func NewService(config *ssh.AuthConfig) *service {
	return &service{
		config: config,
	}
}
