package toolbox

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

//Decoder represents a decoder.
type Decoder interface {
	//Decode  reads and decodes objects from an input stream.
	Decode(v interface{}) error
}

//UnMarshaler represent an struct that can be converted to bytes
type UnMarshaler interface {

	//Unmarshal converts a struct to bytes
	Unmarshal(data []byte) error
}

//DecoderFactory create an decoder for passed in  input stream
type DecoderFactory interface {
	//Create a decoder for passed in io reader
	Create(reader io.Reader) Decoder
}

type jsonDecoderFactory struct{ useNumber bool }

func (d jsonDecoderFactory) Create(reader io.Reader) Decoder {
	decoder := json.NewDecoder(reader)
	if d.useNumber {
		decoder.UseNumber()
	}
	return decoder
}

//NewJSONDecoderFactory create a new JSONDecoderFactory
func NewJSONDecoderFactory() DecoderFactory {
	return &jsonDecoderFactory{}
}

//NewJSONDecoderFactoryWithOption create a new JSONDecoderFactory, it takes useNumber decoder parameter
func NewJSONDecoderFactoryWithOption(useNumber bool) DecoderFactory {
	return &jsonDecoderFactory{useNumber: useNumber}
}

type unMarshalerDecoderFactory struct {
}

func (f *unMarshalerDecoderFactory) Create(reader io.Reader) Decoder {
	return &unMarshalerDecoder{
		reader: reader,
	}
}

type unMarshalerDecoder struct {
	reader   io.Reader
	provider func() UnMarshaler
}

func (d *unMarshalerDecoder) Decode(v interface{}) error {
	bytes, err := ioutil.ReadAll(d.reader)
	if err != nil {
		return fmt.Errorf("Failed to decode %v", err)
	}
	result, casted := v.(UnMarshaler)
	if !casted {
		return fmt.Errorf("Failed to decode - unable cast %T to %s", v, (*UnMarshaler)(nil))
	}
	return result.Unmarshal(bytes)
}

//NewUnMarshalerDecoderFactory returns a decoder factory
func NewUnMarshalerDecoderFactory() DecoderFactory {
	return &unMarshalerDecoderFactory{}
}
