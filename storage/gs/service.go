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
	"github.com/viant/toolbox"
	tstorage "github.com/viant/toolbox/storage"
	"google.golang.org/api/option"
	"os"
	"strings"
	"time"
)

type service struct {
	projectID string
	options   []option.ClientOption
}

func (s *service) NewClient() (*storage.Client, context.Context, error) {
	if s.projectID == "" {
		return nil, nil, fmt.Errorf("project ID was empty, consider setting GOOGLE_CLOUD_PROJECT")
	}
	ctx := context.Background()
	deadline, _ := ctx.Deadline()
	deadline.Add(time.Minute * 30)
	client, err := storage.NewClient(ctx, s.options...)
	if err != nil {
		err = fmt.Errorf("failed to create google storage client:%v", err)
	}
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
func (s *service) Download(object tstorage.Object) (io.ReadCloser, error) {
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
	return client.Bucket(objectInfo.Bucket).
		Object(objectInfo.Name).
		NewReader(ctx)
}

func (s *service) Upload(URL string, reader io.Reader) error {
	return s.UploadWithMode(URL, tstorage.DefaultFileMode, reader)
}

func (s *service) UploadWithMode(URL string, mode os.FileMode, reader io.Reader) error {
	parserURL, err := url.Parse(URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL for uploading: %v, %v", URL, err)
	}
	client, ctx, err := s.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	name := parserURL.Path
	if len(parserURL.Path) > 0 {
		name = parserURL.Path[1:]
	}

	err = s.uploadContent(ctx, client, parserURL, name, reader)
	if toolbox.IsNotFoundError(err) {
		err := client.Bucket(parserURL.Host).Create(ctx, s.projectID, &storage.BucketAttrs{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v, %v", parserURL.Host, err)
		}
		//_, _ = client.Bucket(parserURL.Host).DefaultObjectACL().List(ctx)
		return s.uploadContent(ctx, client, parserURL, name, reader)
	}
	if err != nil {
		return fmt.Errorf("unable upload: %v", err)
	}
	return nil
}

func (s *service) uploadContent(ctx context.Context, client *storage.Client, parserURL *url.URL, name string, reader io.Reader) error {
	writer := client.Bucket(parserURL.Host).
		Object(name).
		NewWriter(ctx)
	expiry := parserURL.Query().Get("expiry")
	if expiry != "" {
		writer.Metadata = map[string]string{
			"Cache-Control": "private, max-age=" + expiry,
		}
	}
	reader, err := updateUploadChecksum(parserURL, reader, writer)
	if _, err = io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to copy to writer during upload:%v", err)
	}
	if err = writer.Close(); err != nil {
		return toolbox.ReclassifyNotFoundIfMatched(err, parserURL.String())
	}
	return nil
}

func updateUploadChecksum(parserURL *url.URL, reader io.Reader, writer *storage.Writer) (io.Reader, error) {
	checksumDisabled := parserURL.Query().Get("disableChecksum") != ""
	updateMD5 := parserURL.Query().Get("disableMD5") == ""
	updateCRC := parserURL.Query().Get("disableCRC32") == ""
	if !(updateCRC || updateMD5) || checksumDisabled {
		return reader, nil
	}

	var err error
	bufferReader, ok := reader.(*bytes.Buffer)
	if !ok {
		content, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read all during upload:%v", err)
		}
		bufferReader = bytes.NewBuffer(content)
	}
	if parserURL.Query().Get("disableMD5") == "" {
		hashReader := bytes.NewBuffer(bufferReader.Bytes())
		h := md5.New()
		_, _ = io.Copy(h, hashReader)
		writer.MD5 = h.Sum(nil)
		hashReader.Reset()
	} else if parserURL.Query().Get("disableCRC32") == "" {
		crc32HashReader := bytes.NewBuffer(bufferReader.Bytes())
		crc32Hash := crc32.New(crc32.MakeTable(crc32.Castagnoli))
		_, _ = io.Copy(crc32Hash, crc32HashReader)
		writer.CRC32C = crc32Hash.Sum32()
		crc32HashReader.Reset()
	}
	return bufferReader, err
}

func (s *service) Register(schema string, service tstorage.Service) error {
	return errors.New("unsupported")
}

func (s *service) Close() error {
	return nil
}

func (s *service) listAll(URL string, result *[]tstorage.Object) error {
	if !strings.HasSuffix(URL, "/") {
		URL += "/"
	}
	objects, err := s.List(URL)
	if err != nil {
		return err
	}
	for _, object := range objects {
		if !object.IsFolder() {
			*result = append(*result, object)
			continue
		}
		if err = s.listAll(object.URL(), result); err != nil {
			return err
		}
	}
	return nil
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
	if object.IsFolder() {
		var objects = []tstorage.Object{}
		err := s.listAll(object.URL(), &objects)
		if err != nil {
			return err
		}
		for _, object := range objects {
			if err := s.Delete(object); err != nil {
				return err
			}
		}
		return nil
	}
	return client.Bucket(objectInfo.Bucket).
		Object(objectInfo.Name).Delete(ctx)
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
func NewService(projectId string, options ...option.ClientOption) *service {
	return &service{
		projectID: projectId,
		options:   options,
	}
}
