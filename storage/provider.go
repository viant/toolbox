package storage

//Provider represetns a service provider
type Provider func(credentialFile string) (Service, error)
