package toolbox

import "time"

const (
	undefined int = iota
	StorageObjectFolderType
	StorageObjectContentType
)

//StorageObject represents a storage object
type StorageObject interface {

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

//AbstractStorageObject represents abstract storage object
type AbstractStorageObject struct {
	StorageObject
	url          string
	objectType   int
	lastModified *time.Time
	size         int64
	Source       interface{}
}

//URL return storage url
func (o *AbstractStorageObject) URL() string {
	return o.url
}

//Type returns storage type  StorageObjectFolderType or StorageObjectContentType
func (o *AbstractStorageObject) Type() int {
	return o.objectType
}

//IsFolder returns true if object is a folder
func (o *AbstractStorageObject) IsFolder() bool {
	return o.objectType == StorageObjectFolderType
}

//IsContent returns true if object is a file
func (o *AbstractStorageObject) IsContent() bool {
	return o.objectType == StorageObjectContentType
}

//LastModified returns last modification time
func (o *AbstractStorageObject) LastModified() *time.Time {
	return o.lastModified
}

//Size returns content size
func (o *AbstractStorageObject) Size() int64 {
	return o.size
}

//Wrap wraps source storage object
func (o *AbstractStorageObject) Wrap(source interface{}) {
	o.Source = source
}

//NewAbstractStorageObject creates a new abstract storage object
func NewAbstractStorageObject(url string, source interface{}, objectType int, lastModified *time.Time, size int64) *AbstractStorageObject {
	return &AbstractStorageObject{
		url:          url,
		Source:       source,
		objectType:   objectType,
		lastModified: lastModified,
		size:         size,
	}
}
