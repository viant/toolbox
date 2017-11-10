package storage

import (
	"io"
	"fmt"
	"github.com/viant/toolbox"
	"path"
	"strings"
	"archive/zip"
)

type CopyHandler func(sourceObject Object, source io.Reader, destinationService Service, destinationURL string) error
type ModificationHandler func(reader io.Reader) (io.Reader, error)

func copy(sourceService Service, sourceURL string, destinationService Service, destinationURL string, modifyContentHandler ModificationHandler, subPath string, copyHandler CopyHandler) error {
	sourceListURL := sourceURL
	if subPath != "" {
		sourceListURL = toolbox.URLPathJoin(sourceURL, subPath)
	}
	objects, err := sourceService.List(sourceListURL)
	var objectRelativePath string
	for _, object := range objects {
		if object.IsFolder() {
			if object.URL() == sourceURL {
				continue
			}
			if subPath != "" && object.URL() == toolbox.URLPathJoin(sourceURL, subPath) {
				continue
			}
		}
		if len(object.URL()) > len(sourceURL) {
			objectRelativePath = object.URL()[len(sourceURL):]
		}
		var destinationObjectURL = destinationURL
		if objectRelativePath != "" {
			destinationObjectURL = toolbox.URLPathJoin(destinationURL, objectRelativePath)
		}

		var reader io.Reader
		if object.IsContent() {
			reader, err = sourceService.Download(object)
			if err != nil {
				err = fmt.Errorf("Unable download, %v -> %v, %v", object.URL(), destinationObjectURL, err)
				return err
			}

			if modifyContentHandler != nil {
				reader, err = modifyContentHandler(reader)
				if err != nil {
					err = fmt.Errorf("Unable modify content, %v %v %v", object.URL(), destinationObjectURL, err)
					return err
				}
			}
			destinationObject, err := destinationService.StorageObject(destinationObjectURL)
			if (subPath == "" && destinationObject != nil && destinationObject.IsFolder()) {
				_, file := path.Split(object.URL())
				destinationObjectURL = toolbox.URLPathJoin(destinationObjectURL, file)
			}

			err = copyHandler(object, reader, destinationService, destinationObjectURL)
			if err != nil {
				return nil
			}

		} else {

			err = copy(sourceService, sourceURL, destinationService, destinationURL, modifyContentHandler, objectRelativePath, copyHandler)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copySourceToDestination(sourceObject Object, reader io.Reader, destinationService Service, destinationURL string) error {
	err := destinationService.Upload(destinationURL, reader)
	if err != nil {
		err = fmt.Errorf("Unable upload, %v %v %v", sourceObject.URL(), destinationURL, err)
	}
	return err
}

func addPathIfNeeded(directories map[string]bool, path string, archive zip.Writer) {
	if path == "" {
		return
	}
	if _, has := directories[path]; has {
		return
	}

}

func getArchiveCopyHandler(archive zip.Writer, parentURL string) CopyHandler {
	var directories = make(map[string]bool)
	return func(sourceObject Object, reader io.Reader, destinationService Service, destinationURL string) error {
		var _, relativePath = toolbox.URLSplit(destinationURL);
		if destinationURL != parentURL {
			relativePath = strings.Replace(destinationURL, parentURL, "", 1)
			var parent, _ = path.Split(relativePath)
			addPathIfNeeded(directories, parent, archive)
		}
		header, err := zip.FileInfoHeader(sourceObject.FileInfo())
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		header.Name = relativePath
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, reader)
		return err
	}
}

//Copy downloads objects from source URL to upload them to destination URL.
func Copy(sourceService Service, sourceURL string, destinationService Service, destinationURL string, modifyContentHandler ModificationHandler, copyHandler CopyHandler) (err error) {
	if copyHandler == nil {
		copyHandler = copySourceToDestination
	}
	err = copy(sourceService, sourceURL, destinationService, destinationURL, modifyContentHandler, "", copyHandler)
	if err != nil {
		err = fmt.Errorf("Failed to copy %v -> %v: %v", sourceURL, destinationURL, err)
	}
	return err
}
