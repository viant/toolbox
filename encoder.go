package toolbox

import (
	"encoding/json"
	"io"
)

//Encoder writes an instance to output stream
type Encoder interface {

	//Encode encodes  an instance to output stream
	Encode(object interface{}) error
}

//EncoderFactory create an encoder for an output stream
type EncoderFactory interface {
	//Create creates an encoder for an output stream
	Create(writer io.Writer) Encoder
}

type jsonEncoderFactory struct{}

func (e jsonEncoderFactory) Create(writer io.Writer) Encoder {
	return json.NewEncoder(writer)
}

//NewJSONEncoderFactory creates new NewJSONEncoderFactory
func NewJSONEncoderFactory() EncoderFactory {
	return &jsonEncoderFactory{}
}
