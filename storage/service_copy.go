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


func urlPath(URL string) string {
	var result = URL
	schemaPosition:= strings.Index(URL, "://")
	if schemaPosition != -1 {
		result = string(URL[schemaPosition+3:])
	}
	pathRoot := strings.Index(result, "/")
	if pathRoot > 0 {
		result= string(result[pathRoot:])
	}
	return result
}

func copy(sourceService Service, sourceURL string, destinationService Service, destinationURL string, modifyContentHandler ModificationHandler, subPath string, copyHandler CopyHandler) error {
	sourceListURL := sourceURL
	if subPath != "" {
		sourceListURL = toolbox.URLPathJoin(sourceURL, subPath)
	}
	objects, err := sourceService.List(sourceListURL)
	var objectRelativePath string
	sourceURLPath := urlPath(sourceURL)
	for _, object := range objects {
		var objectURLPath = urlPath(object.URL())
		if object.IsFolder() {

			if sourceURLPath == objectURLPath {
				continue
			}
			if subPath != "" && objectURLPath== toolbox.URLPathJoin(sourceURLPath, subPath) {
				continue
			}
		}
		if len(objectURLPath) > len(sourceURLPath) {
			objectRelativePath = objectURLPath[len(sourceURLPath):]
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

			if subPath == "" {
				_, sourceName := path.Split(object.URL())
				_, destinationName := path.Split(destinationURL)
				if strings.HasSuffix(destinationObjectURL, "/") {
					destinationObjectURL = toolbox.URLPathJoin(destinationObjectURL, sourceName)
				} else {
					destinationObject, _ := destinationService.StorageObject(destinationObjectURL)
					if (destinationObject != nil && destinationObject.IsFolder()) {
						destinationObjectURL = toolbox.URLPathJoin(destinationObjectURL, sourceName)
					} else if destinationName != sourceName   {
						if ! strings.Contains(destinationName, ".") {
							destinationObjectURL = toolbox.URLPathJoin(destinationObjectURL, sourceName)
						}

					}
				}
			}

			err = copyHandler(object, reader, destinationService, destinationObjectURL)
			if err != nil {
				return err
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
