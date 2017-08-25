package storage

type Provider func(credentialFile string) (Service, error)
type StorageProvider struct {
	Registry map[string]Provider
}

func (p *StorageProvider) Get(namespace string) func(credentialFile string) (Service, error) {
	return p.Registry[namespace]
}

var storageProvider *StorageProvider

func NewStorageProvider() *StorageProvider {
	if storageProvider != nil {
		return storageProvider
	}
	storageProvider = &StorageProvider{
		Registry: make(map[string]Provider),
	}
	return storageProvider
}
