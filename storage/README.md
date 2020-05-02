## Storage API

Deprecated - please use https://github.com/viant/afs API instead

This API provides unified way of accessing any storage system. 

It comes with the following implementation so far:



<a name="import></a>


| URL Scheme | Description | Pacakge |
|-----|-----|-----|
|file | Local file system | github.com/viant/toolbox/storage |
|https | HTTP/s based system | github.com/viant/toolbox/storage |
|mem | Memory file system | github.com/viant/toolbox/storage |
|scp | SCP/SSH base systm | github.com/viant/toolbox/storage/scp |
|s3 |Amazon Web Service S3| github.com/viant/toolbox/storage/aws |
|gs | Google Storage | github.com/viant/toolbox/storage/gs |



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

	//Wrap wraps source storage object
	Wrap(source interface{})

	//Unwrap unwraps source storage object into provided target.
	Unwrap(target interface{}) error

	FileInfo() os.FileInfo
}

```

**Usage:** 


```go
    import (
    	"github.com/viant/toolbox/storage"
    	_ "github.com/viant/toolbox/storage/gs"
    	_ "github.com/viant/toolbox/storage/s3"
	
    )

    destinationURL := "gs://myBucket/set1/content.gz"
    destinationCredentialFile = "gs-secret.json"
	storageService, err := storage.NewServiceForURL(destinationURL, destinationCredentialFile)

    provider := storage.Registry().Get("s3")
    storageS3Service, err := provider("aws-secret.json")

```
