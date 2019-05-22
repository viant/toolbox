package storage

type registry struct {
	Registry map[string]Provider
}

func (p *registry) Get(namespace string) func(credentialFile string) (Service, error) {
	return p.Registry[namespace]
}

var registrySingleton *registry

//Registry returns new provider
func Registry() *registry {
	if registrySingleton != nil {
		return registrySingleton
	}
	registrySingleton = &registry{
		Registry: make(map[string]Provider),
	}
	return registrySingleton
}
