package scp

import (
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/storage"
	"path"
	"strings"
	"time"
)

type object struct {
	*storage.AbstractObject
	source           interface{}
	url              string
	name             string
	owner            string
	group            string
	size             string
	modificationTime time.Time
	day              string
	month            string
	year             string
	hour             string
	isDirectory      bool
	permission       string
}

//URL return storage url
func (i *object) URL() string {
	return path.Join(i.url, i.name)
}

//Type returns storage type  StorageObjectFolderType or StorageObjectContentType
func (i *object) Type() int {
	if strings.Contains(i.permission, "d") {
		return storage.StorageObjectFolderType
	}
	return storage.StorageObjectContentType
}

//IsFolder returns true if object is a folder
func (i *object) IsFolder() bool {
	return i.Type() == storage.StorageObjectFolderType
}

//IsContent returns true if object is a file
func (i *object) IsContent() bool {
	return i.Type() == storage.StorageObjectContentType
}

//LastModified returns last modification time
func (i *object) LastModified() *time.Time {
	dateTime := i.year + " " + i.month + " " + i.day + " " + i.hour
	layout := toolbox.DateFormatToLayout("yyyy MMM ddd HH:mm:s")
	result, err := time.Parse(layout, dateTime)
	if err != nil {
		return nil
	}
	return &result

}

//Size returns content size
func (i *object) Size() int64 {
	return int64(toolbox.AsInt(i.size))
}

//Wrap wraps source storage object
func (i *object) Wrap(source interface{}) {
	i.source = source
}

//Unwrap unwraps source storage object into provided target.
func (i *object) Unwrap(target interface{}) error {
	if result, ok := target.(**object); ok {
		*result = i
	}
	return nil
}

//newObject creates a new gc storage object
func newStorageObject(url string, objectType int, source interface{}, lastModified *time.Time, size int64) storage.Object {
	abstract := storage.NewAbstractStorageObject(url, source, objectType, lastModified, size)
	result := &object{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}
