package gs

import (
	"cloud.google.com/go/storage"
	"fmt"
	tstorage "github.com/viant/toolbox/storage"
	"os"
)

type object struct {
	*tstorage.AbstractObject
}

func (o *object) Unwrap(target interface{}) error {
	if fileInfo, casted := target.(**storage.ObjectAttrs); casted {
		source, ok := o.Source.(*storage.ObjectAttrs)
		if !ok {
			return fmt.Errorf("failed to case %T into %T", o.Source, target)
		}
		*fileInfo = source
		return nil
	}
	return fmt.Errorf("unsuported target %T", target)
}

//newObject creates a new gc storage object
func newStorageObject(url string, source interface{}, fileInfo os.FileInfo) tstorage.Object {
	abstract := tstorage.NewAbstractStorageObject(url, source, fileInfo)
	result := &object{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}
