package gs

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"net/url"
	"cloud.google.com/go/storage"
	tstorage "github.com/viant/toolbox/storage"
	"google.golang.org/api/option"
	"github.com/viant/toolbox"
)


type service struct {
	options []option.ClientOption
}

func (s *service) NewClient() (*storage.Client, context.Context, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, s.options...)
	return client, ctx, err
}

//List returns a list of object for supplied url
func (s *service) List(URL string) ([]tstorage.Object, error) {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	client, ctx, err := s.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var query = &storage.Query{
		Delimiter: "/",
	}
	if len(parsedUrl.Path) > 0 {
		query.Prefix = parsedUrl.Path[1:]
	}
	responseIterator := client.Bucket(parsedUrl.Host).Objects(ctx, query)
	var result = make([]tstorage.Object, 0)
	for obj, err := responseIterator.Next(); err == nil; obj, err = responseIterator.Next() {
		objectURL := "gs://" + parsedUrl.Host + "/" + obj.Prefix + obj.Name
		var fileMode, _ = tstorage.NewFileMode("-rw-rw-rw-")
		if obj.Prefix != "" {
			fileMode, _ = tstorage.NewFileMode("drw-rw-rw-")
		}
		var _, name = toolbox.URLSplit(objectURL)
		var fileInfo = tstorage.NewFileInfo(name, obj.Size, fileMode, obj.Updated, fileMode.IsDir())
		var object = newStorageObject(objectURL, obj, fileInfo)
		result = append(result, object)
	}
	return result, err
}

func (s *service) Exists(URL string) (bool, error) {
	objects, err := s.List(URL)
	if err != nil {
		return false, err
	}
	return len(objects) > 0, nil
}

func (s *service) StorageObject(URL string) (tstorage.Object, error) {
	objects, err := s.List(URL)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("Not found %v", URL)
	}
	return objects[0], nil
}

//Download returns reader for downloaded storage object
func (s *service) Download(object tstorage.Object) (io.Reader, error) {
	client, ctx, err := s.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	objectInfo := &storage.ObjectAttrs{}
	err = object.Unwrap(&objectInfo)
	if err != nil {
		return nil, err
	}
	reader, err := client.Bucket(objectInfo.Bucket).
		Object(objectInfo.Name).
		NewReader(ctx)

	if err != nil {
		return nil, err
	}
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(content), err
}

//Upload uploads provided reader content for supplied url.
func (s *service) Upload(URL string, reader io.Reader) error {
	parsedUrl, err := url.Parse(URL)
	if err != nil {
		return err
	}
	client, ctx, err := s.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	name := parsedUrl.Path
	if len(parsedUrl.Path) > 0 {
		name = parsedUrl.Path[1:]
	}
	writer := client.Bucket(parsedUrl.Host).
		Object(name).
		NewWriter(ctx)

	expiry := parsedUrl.Query().Get("expiry")
	if expiry != "" {
		writer.Metadata = map[string]string{
			"Cache-Control": "private, max-age=" + expiry,
		}
	}
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	contentReader := bytes.NewBuffer(content)

	if parsedUrl.Query().Get("disableMD5") == "" {
		hashReader := bytes.NewBuffer(content)
		h := md5.New()
		io.Copy(h, hashReader)
		writer.MD5 = h.Sum(nil)
	}

	if parsedUrl.Query().Get("disableCRC32") == "" {
		crc32HashReader := bytes.NewBuffer(content)
		crc32Hash := crc32.New(crc32.MakeTable(crc32.Castagnoli))
		io.Copy(crc32Hash, crc32HashReader)
		writer.CRC32C = crc32Hash.Sum32()
	}

	if _, err = io.Copy(writer, contentReader); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	return err
}

func (s *service) Register(schema string, service tstorage.Service) error {
	return errors.New("unsupported")
}

//Delete removes passed in storage object
func (s *service) Delete(object tstorage.Object) error {
	client, ctx, err := s.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	objectInfo := &storage.ObjectAttrs{}
	err = object.Unwrap(&objectInfo)
	if err != nil {
		return err
	}
	return client.Bucket(objectInfo.Bucket).
		Object(objectInfo.Name).Delete(ctx)
}

//NewService create a new gc storage service
func NewService(options ...option.ClientOption) *service {
	return &service{
		options: options,
	}
}
