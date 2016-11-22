package toolbox

import (
	"encoding/json"
	"fmt"
	"io"
)

//Encoder writes an instance to output stream
type Encoder interface {
	//Encode encodes  an instance to output stream
	Encode(object interface{}) error
}

//Marshaler represents byte to object converter
type Marshaler interface {
	//Marshal converts bytes to attributes of owner struct
	Marshal() (data []byte, err error)
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

type marshalerEncoderFactory struct {
}

func (f *marshalerEncoderFactory) Create(writer io.Writer) Encoder {
	return &marshalerEncoder{writer: writer}
}

type marshalerEncoder struct {
	writer io.Writer
}

func (e *marshalerEncoder) Encode(v interface{}) error {
	result, casted := v.(Marshaler)
	if !casted {
		return fmt.Errorf("Failed to decode - unable cast %T to %s", v, (*Marshaler)(nil))
	}
	bytes, err := result.Marshal()
	if err != nil {
		return err
	}

	for i := 0; i < len(bytes); i++ {
		count, err := e.writer.Write(bytes)
		if err != nil {
			return err
		}
		if count == len(bytes) {
			break
		}
		bytes = bytes[count:]
	}
	return nil
}

//NewMarshalerEncoderFactory create a new encoder factory for marsheler struct
func NewMarshalerEncoderFactory() EncoderFactory {
	return &marshalerEncoderFactory{}
}
