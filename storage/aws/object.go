package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/viant/toolbox/storage"
	"time"
)

type object struct {
	*storage.AbstractObject
}

func (o *object) Unwrap(target interface{}) error {
	if commonPrefix, casted := target.(**s3.CommonPrefix); casted {
		source, ok := o.Source.(*s3.CommonPrefix)
		if !ok {
			return fmt.Errorf("Failed to case %T into %T", o.Source, target)
		}
		*commonPrefix = source
	}
	if commonPrefix, casted := target.(**s3.Object); casted {
		source, ok := o.Source.(*s3.Object)
		if !ok {
			return fmt.Errorf("Failed to case %T into %T", o.Source, target)
		}
		*commonPrefix = source
	}
	return fmt.Errorf("unsuported target %T", target)
}

//newObject creates a new aws storage object
func newObject(url string, objectType int, source interface{}, lastModified *time.Time, size int64) storage.Object {
	abstract := storage.NewAbstractStorageObject(url, source, objectType, lastModified, size)
	result := &object{
		AbstractObject: abstract,
	}
	result.AbstractObject.Object = result
	return result
}
