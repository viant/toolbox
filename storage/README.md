#Storage API



This API provides unified way of accessing any storage system. 

It comes with the following implementation so far:

1) Local file system
2) SCP/SSH
3) Memory file system
4) Google Storage
5) Amazon Web Service S3
6) HTTP/S


```go
type Service interface {
	//List returns a list of object for supplied url
	List(URL string) ([]Object, error)

	//Exists returns true if resource exists
	Exists(URL string) (bool, error)

	//Object returns a Object for supplied url
	StorageObject(URL string) (Object, error)

	//Download returns reader for downloaded storage object
	Download(object Object) (io.ReadCloser, error)

	//Upload uploads provided reader content for supplied storage object.
	Upload(URL string, reader io.Reader) error

	//Delete removes passed in storage object
	Delete(object Object) error

	//Register register schema with provided service
	Register(schema string, service Service) error

	//Closes storage service
	Close() error
}

```

**Usage** 


```go
    import (
    	"github.com/viant/toolbox/storage"
    	_ "github.com/viant/toolbox/storage/gs"
	
    )

    destinationURL := "gs://myBucket/set1/content.gz"
    destinationCredentialFile = "gs-secret.json"
	storageService, err := storage.NewServiceForURL(destinationURL, destinationCredentialFile)


```
