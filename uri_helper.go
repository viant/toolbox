package toolbox

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

//FileSchema file://
var FileSchema = "file://"

//ExtractMimeType extracts mime type by extension
func ExtractMimeType(file string) string {
	extension := path.Ext(file)
	if len(extension) > 1 {
		extension = extension[1:]
	}

	if mimeType, ok := FileExtensionMimeType[extension]; ok {
		return mimeType
	}
	return "text/plain"
}

//OpenReaderFromURL opens a reader from URL
// to indicate a relative path from the working directory
// the schema must start with "file://.../". "..." will be replaced by
// the working directory. Note that you can not have a relative path
// in the middle of the uri. for instance, "file:///local/../path
// is not valid
func OpenReaderFromURL(rawURL string) (io.ReadCloser, string, error) {
	//this indicates a relative path


	if strings.HasPrefix(rawURL, "file://.../") {
		baseDirectory, _ := os.Getwd()
		rawURL = strings.Replace(rawURL, "...", baseDirectory, 1)
	}

	var url, err = url.Parse(rawURL)
	if err != nil {
		return nil, "", err
	}
	switch url.Scheme {
	case "http", "https":
		response, err := http.Get(rawURL)
		if err != nil {
			return nil, "", err
		}
		mimeType := response.Header.Get("Content-Type")
		return response.Body, mimeType, nil
	case "file":
		file, err := OpenFile(url.Path)
		if err != nil {
			return nil, "", fmt.Errorf("failed to open file: %v due to %v", rawURL, err.Error())
		}
		return file, ExtractMimeType(url.Path), nil
	}
	return nil, "", fmt.Errorf("Unsupprted url.Scheme: %v on %v", url.Scheme, rawURL)
}

//FileFromURL returns file path from passed in URL.
func FileFromURL(fileURL string) (string, error) {
	var url, err = url.Parse(fileURL)
	if err != nil {
		return "", err
	}
	switch url.Scheme {
	case "file":
		return url.Path, nil
	}
	return "", fmt.Errorf("Unsupprted url.Scheme: %v on %v", url.Scheme, fileURL)
}

//OpenURL opens passed in url as file, or error. Only file:// scheme is supported
func OpenURL(fileURL string, flag int, permissions os.FileMode) (*os.File, error) {
	filePath, err := FileFromURL(fileURL)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filePath, flag, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v due to %v", fileURL, err.Error())
	}
	return file, nil
}

//QueryIntValue returns query value for passed in url's name or default value
func QueryIntValue(u *url.URL, name string, defaultValue int) int {
	value := u.Query().Get(name)
	if value == "" {
		return defaultValue
	}

	result := AsInt(value)
	if result != 0 {
		return result
	}
	return defaultValue
}

//QueryBoolValue returns query value for passed in url's name or default value
func QueryBoolValue(u *url.URL, name string, defaultValue bool) bool {
	value := u.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	return AsBoolean(value)
}

//QueryValue returns query value for passed in url's name or default value
func QueryValue(u *url.URL, name, defaultValue string) string {
	value := u.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	return value
}
