package storage

import (
	"os"
)

const (
	undefined                int = iota
	StorageObjectFolderType      //folder type
	StorageObjectContentType     //file type
)

//Object represents a storage object
type Object interface {
	//URL return storage url
	URL() string
	//Type returns storage type either folder or file
	Type() int
	//IsFolder returns true if object is a folder
	IsFolder() bool
	//IsContent returns true if object is a file
	IsContent() bool
	//Wrap wraps source storage object
	Wrap(source interface{})
	//Unwrap unwraps source storage object into provided target.
	Unwrap(target interface{}) error
	FileInfo() os.FileInfo
}

//AbstractObject represents abstract storage object
type AbstractObject struct {
	Object
	url        string
	objectType int
	Source     interface{}
	fileInfo   os.FileInfo
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

//Wrap wraps source storage object
func (o *AbstractObject) Wrap(source interface{}) {
	o.Source = source
}

func (o *AbstractObject) FileInfo() os.FileInfo {
	return o.fileInfo
}

//NewAbstractStorageObject creates a new abstract storage object
func NewAbstractStorageObject(url string, source interface{}, fileInfo os.FileInfo) *AbstractObject {
	var result = &AbstractObject{
		url:        url,
		Source:     source,
		fileInfo:   fileInfo,
		objectType: StorageObjectContentType,
	}
	if fileInfo.IsDir() {
		result.objectType = StorageObjectFolderType
	}
	return result
}
