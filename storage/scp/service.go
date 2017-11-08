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

	"path"
	"github.com/viant/toolbox/cred"
	"net/url"
	"time"
)

const defaultSSHPort = 22
const verificationSizeThreshold = 10 * 1024 * 1024

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

const (
	fileIsoInfoPermission = iota
	_
	fileIsoInfoOwner
	fileIsoInfoGroup
	fileIsoInfoSize
	fileIsoDate
	fileIsoTime
	fileIsoTimezone
	fileIsoInfoName
)



type service struct {
	config *cred.Config

}

func (s *service) runCommand(session *ssh.MultiCommandSession, URL string, command string) (string, error) {
	output, _ := session.Run(command, 0, "$ ", "usage", "No such file or directory")
	return toolbox.AsString(output), nil
}


func (s *service) canListWithTimeStyle(session *ssh.MultiCommandSession, URL string) (bool) {
	return session.KernelName != "darwin"
}


func normalizeFileInfoOutput(lines string) string {
	var result = make([]string, 0)
	for _, line := range strings.Split(lines, "\n") {
		line = strings.Replace(line, "\r", "", 1)
		if strings.HasPrefix(strings.ToLower(line), "total") {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}


func  (s *service) getClient(parsedURL *url.URL) (*ssh.Client, error) {
	port := toolbox.AsInt(parsedURL.Port())
	if port == 0 {
		port = 22
	}
	return ssh.NewClient(parsedURL.Hostname(), toolbox.AsInt(port), s.config)
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


	var urlPath = strings.Replace(parsedUrl.Path, "//", "/", len(parsedUrl.Path))
	var result = make([]storage.Object, 0)

	canListWithTimeStyle := s.canListWithTimeStyle(session, URL)
	var lsCommand = "ls -ltr"
	if canListWithTimeStyle {
		lsCommand += " --time-style=full-iso"
	} else {
		lsCommand +="T"
	}

	output, err := s.runCommand(session, URL, lsCommand + " "+parsedUrl.Path)
	stdout := normalizeFileInfoOutput(string(output))
	if strings.Contains(stdout, "No such file or directory") {
		return result, nil
	}
	var fileNameFilter = ""


	if err == nil && stdout == "" {
		parent, fileName := path.Split(urlPath )
		fileNameFilter = fileName
		output, err = s.runCommand(session, URL, lsCommand + " "+parent+" | grep "+fileName)
	}
	if err != nil {
		return nil, err
	}

	stdout = normalizeFileInfoOutput(string(output))
	for _, line := range strings.Split(stdout, "\n") {
		fileInfo := ExtractFileInfo(line, canListWithTimeStyle)
		if fileInfo == nil {
			continue
		}
		if fileInfo.name == "" {
			continue
		}
		fileInfo.url = URL
		if fileNameFilter == "" || fileNameFilter == fileInfo.name {
			result = append(result, fileInfo)
		}
	}
	return result, nil
}



//file info with iso -rw-r--r-- 1 awitas awitas 2002 2017-11-04 22:29:33.363458941 +0000 aerospikeciads_aerospike.conf
//file info without iso // -rw-r--r--  1 awitas  1742120565   414 Jun  8 14:14:08 2017 id_rsa.pub

func ExtractFileInfo(line string, isoTimeStyle bool) *object {
	fragmentCount := 0
	fileInfo := &object{}
	if strings.TrimSpace(line) == "" {
		return nil
	}
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

		if isoTimeStyle {
			switch fragmentCount {
			case fileIsoInfoPermission:
				fileInfo.permission += aChar
			case fileIsoInfoOwner:
				fileInfo.owner += aChar
			case fileIsoInfoGroup:
				fileInfo.group += aChar
			case fileIsoInfoSize:
				fileInfo.size += aChar
			case fileIsoDate:
				fileInfo.date += aChar
			case fileIsoTime:
				fileInfo.time += aChar
			case fileIsoTimezone:
				fileInfo.timezone += aChar
			case fileIsoInfoName:
				fileInfo.name += aChar
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

	output, err := s.runCommand(session, URL, "ls -ltr "+parsedUrl.Path)
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
		return nil, fmt.Errorf("Not found %v", URL)
	}
	if len(objects) > 1 || (len(objects)  == 1  && objects[0].URL() != URL) {
		var now = time.Now()
		fileInfo := &object{
			permission:"drwrwrw",
			modificationTime:&now,
		}
		_, fileInfo.name = toolbox.URLSplit(URL)
		fileInfo.url = URL
		return fileInfo, nil
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

	if  verificationSizeThreshold < len(content) {
		//download verification (as sometimes scp failed) with one retry
		if int(object.Size()) != len(content) {
			content, err = client.Download(parsedUrl.Path)
			if err != nil {
				return nil, err
			}
			if int(object.Size()) != len(content) {
				return nil, fmt.Errorf("Faled to download from %v,  object size was: %v, but scp download was %v", object.URL(), object.Size(), len(content))
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
	client, err := ssh.NewClient(parsedUrl.Hostname(), toolbox.AsInt(port), s.config)
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
		return  fmt.Errorf("Failed to upload: %v %v", URL,  err)
	}



	if  verificationSizeThreshold < len(content) {
		object, err := s.StorageObject(URL)
		if err != nil {
			return  fmt.Errorf("Failed to get upload object  %v for verification: %v", URL, err)
		}
		if int(object.Size()) != len(content) {
			err = client.Upload(parsedUrl.Path, content)
			object, err = s.StorageObject(URL)
			if err != nil {
				return  err
			}
			if int(object.Size()) != len(content) {
				return fmt.Errorf("Failed to upload to %v, actual size was:%v,  but uploaded size was  ", URL, len(content), int(object.Size()))
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
func NewService(config *cred.Config) *service {
	return &service{
		config: config,
	}
}
