package storage

import (
	"fmt"
	"github.com/viant/toolbox"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type StorageMapping struct {
	SourceURL        string
	SourceCredential string
	DestinationURI   string
	TargetFile       string
	TargetPackage    string
}

//GenerateStorageCode create a *.go files with statically scanned content from source URL.
func GenerateStorageCode(mappings ...*StorageMapping) error {
	destinationService := NewMemoryService()
	for _, mapping := range mappings {

		sourceService, err := NewServiceForURL(mapping.SourceURL, mapping.SourceCredential)
		if err != nil {
			return err
		}
		handler, writer, err := NewStorageMapperHandler(mapping.TargetFile, mapping.TargetPackage)
		if err != nil {
			return err
		}
		defer writer.Close()
		destinationURL := "mem://" + mapping.DestinationURI
		err = copyStorageContent(sourceService, mapping.SourceURL, destinationService, destinationURL, nil, "", handler)
		if err != nil {
			return err
		}
	}
	return nil
}

//NewStorageMapperHandler creates a template handler for generating go file that write static content into memory service.
func NewStorageMapperHandler(filename, pkg string) (CopyHandler, io.WriteCloser, error) {
	toolbox.RemoveFileIfExist(filename)
	writer, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}
	template := &templateWriter{writer}
	template.Init(pkg)
	return func(sourceObject Object, source io.Reader, destinationService Service, destinationURL string) error {
		content, err := ioutil.ReadAll(source)
		if err != nil {
			return err
		}
		template.WriteStorageContent(destinationURL, content)
		return nil
	}, template, nil
}

type templateWriter struct {
	io.WriteCloser
}

func (t *templateWriter) Init(pkg string) error {
	var begin = `package %v

import (
	"bytes"
	"github.com/viant/toolbox/storage"
	"log"
)

func init() {
	var memStorage = storage.NewMemoryService();
`
	_, err := t.Write([]byte(fmt.Sprintf(begin, pkg)))
	return err
}

func (t *templateWriter) WriteStorageContent(URL string, content []byte) error {
	var contentReader = fmt.Sprintf("bytes.NewReader([]byte(`%s`))", content)
	if !toolbox.IsASCIIText(contentReader) {
		var byteArray = make([]string, 0)
		for _, b := range content {
			byteArray = append(byteArray, fmt.Sprintf("%d", b))
		}
		contentReader = fmt.Sprintf("bytes.NewReader([]byte{%v})", strings.Join(byteArray, ","))
	}
	var payload = `	{
		err := memStorage.Upload("%v", %v)
		if err != nil {
			log.Printf("failed to upload: %v %v", err)
		}
	}
`
	payload = fmt.Sprintf(payload, URL, contentReader, URL, "%v")
	_, err := t.Write([]byte(payload))
	return err
}

func (t *templateWriter) Close() error {
	var end = "}\n"
	_, err := t.Write([]byte(end))
	t.WriteCloser.Close()
	return err
}
