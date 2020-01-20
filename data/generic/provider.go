package dynamic


//Provider provides shares meta data across all dynamic types
type Provider struct {
	Meta
}

//NewObject creates a slice
func (p *Provider) NewSlice() *Slice {
	return &Slice{_data: [][]interface{}{}, meta:&p.Meta}
}

//NewObject creates an object
func (p *Provider) NewObject() *Object {
	return &Object{_data:[]interface{}{}, meta:&p.Meta}
}


//NewProvider creates provider
func NewProvider() *Provider {
	return &Provider{Meta:*NewMeta()}
}
