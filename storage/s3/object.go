package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/viant/toolbox/storage"
	"os"
)

type object struct {
	*storage.AbstractObject
}

func (o *object) Unwrap(target interface{}) error {
	if commonPrefix, casted := target.(**s3.CommonPrefix); casted {
		source, ok := o.Source.(*s3.CommonPrefix)
		if !ok {
			return fmt.Errorf("failed to case %T into %T", o.Source, target)
		}
		*commonPrefix = source
	}
	if commonPrefix, casted := target.(**s3.Object); casted {
		source, ok := o.Source.(*s3.Object)
		if !ok {
			return fmt.Errorf("failed to case %T into %T", o.Source, target)
		}
		*commonPrefix = source
	}

	return fmt.Errorf("unsuported target %T", target)
}

//newStorageObject creates a new aws storage object
func newStorageObject(url string, source interface{}, fileInfo os.FileInfo) storage.Object {
	abstract := storage.NewAbstractStorageObject(url, source, fileInfo)
	result := &object{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}
