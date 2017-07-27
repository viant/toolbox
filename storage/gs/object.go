package gs

import (
	"cloud.google.com/go/storage"
	"fmt"
	tstorage "github.com/viant/toolbox/storage"
	"time"
)

type object struct {
	*tstorage.AbstractObject
}

func (o *object) Unwrap(target interface{}) error {
	if fileInfo, casted := target.(**storage.ObjectAttrs); casted {
		source, ok := o.Source.(*storage.ObjectAttrs)
		if !ok {
			return fmt.Errorf("Failed to case %T into %T", o.Source, target)
		}
		*fileInfo = source
		return nil
	}
	return fmt.Errorf("unsuported target %T", target)
}

//newObject creates a new gc storage object
func newStorageObject(url string, objectType int, source interface{}, lastModified *time.Time, size int64) tstorage.Object {
	abstract := tstorage.NewAbstractStorageObject(url, source, objectType, lastModified, size)
	result := &object{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}
