package storage

import "time"

const (
	undefined                int = iota
	StorageObjectFolderType      //folder type
	StorageObjectContentType     //file type
)

//Object represents a storage object
type Object interface {

	//URL return storage url
	URL() string

	//Type returns storage type  StorageObjectFolderType or StorageObjectContentType
	Type() int

	//IsFolder returns true if object is a folder
	IsFolder() bool

	//IsContent returns true if object is a file
	IsContent() bool

	//LastModified returns last modification time
	LastModified() *time.Time

	//Size returns content size
	Size() int64

	//Wrap wraps source storage object
	Wrap(source interface{})

	//Unwrap unwraps source storage object into provided target.
	Unwrap(target interface{}) error
}

//AbstractObject represents abstract storage object
type AbstractObject struct {
	Object
	url          string
	objectType   int
	lastModified *time.Time
	size         int64
	Source       interface{}
}

//URL return storage url
func (o *AbstractObject) URL() string {
	return o.url
}

//Type returns storage type  StorageObjectFolderType or StorageObjectContentType
func (o *AbstractObject) Type() int {
	return o.objectType
}

//IsFolder returns true if object is a folder
func (o *AbstractObject) IsFolder() bool {
	return o.objectType == StorageObjectFolderType
}

//IsContent returns true if object is a file
func (o *AbstractObject) IsContent() bool {
	return o.objectType == StorageObjectContentType
}

//LastModified returns last modification time
func (o *AbstractObject) LastModified() *time.Time {
	return o.lastModified
}

//Size returns content size
func (o *AbstractObject) Size() int64 {
	return o.size
}

//Wrap wraps source storage object
func (o *AbstractObject) Wrap(source interface{}) {
	o.Source = source
}

//NewAbstractStorageObject creates a new abstract storage object
func NewAbstractStorageObject(url string, source interface{}, objectType int, lastModified *time.Time, size int64) *AbstractObject {
	return &AbstractObject{
		url:          url,
		Source:       source,
		objectType:   objectType,
		lastModified: lastModified,
		size:         size,
	}
}
