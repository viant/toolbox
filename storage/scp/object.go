package scp

import (
	"github.com/viant/toolbox/storage"
	"os"
)

type object struct {
	*storage.AbstractObject
	source     interface{}
	permission string
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
func newStorageObject(URL string, source interface{}, fileInfo os.FileInfo) storage.Object {
	abstract := storage.NewAbstractStorageObject(URL, source, fileInfo)
	result := &object{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}
