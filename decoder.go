package toolbox

import (
	"encoding/json"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"strings"
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
		return fmt.Errorf("failed to decode %v", err)
	}
	result, casted := v.(UnMarshaler)
	if !casted {
		return fmt.Errorf("failed to decode - unable cast %T to %s", v, (*UnMarshaler)(nil))
	}
	return result.Unmarshal(bytes)
}

//NewUnMarshalerDecoderFactory returns a decoder factory
func NewUnMarshalerDecoderFactory() DecoderFactory {
	return &unMarshalerDecoderFactory{}
}

//DelimitedRecord represents a delimited record
type DelimitedRecord struct {
	Columns   []string
	Delimiter string
	Record    map[string]interface{}
}

//IsEmpty returns true if all values are empty or null
func (r *DelimitedRecord) IsEmpty() bool {
	var result = true
	for _, value := range r.Record {
		if value == nil {
			continue
		}
		if AsString(value) == "" || AsString(value) == "<nil>" {
			continue
		}
		return false
	}
	return result
}

type delimiterDecoder struct {
	reader io.Reader
}

func (d *delimiterDecoder) Decode(target interface{}) error {
	delimitedRecord, ok := target.(*DelimitedRecord)
	if !ok {
		return fmt.Errorf("Invalid target type, expected %T but had %T", &DelimitedRecord{}, target)
	}
	if delimitedRecord.Record == nil {
		delimitedRecord.Record = make(map[string]interface{})
	}

	var isInDoubleQuote = false
	var index = 0
	var value = ""
	var delimiter = delimitedRecord.Delimiter
	payload, err := ioutil.ReadAll(d.reader)
	if err != nil {
		return err
	}
	encoded := string(payload)
	hasColumns := len(delimitedRecord.Columns) > 0
	if !hasColumns {
		delimitedRecord.Columns = make([]string, 0)
	}
	for i := 0; i < len(encoded); i++ {
		aChar := string(encoded[i : i+1])
		nextChar := ""
		if i+2 < len(encoded) {
			nextChar = encoded[i+1 : i+2]
		}
		if isInDoubleQuote && ((aChar == "\\" || aChar == "\"") && i+2 < len(encoded)) {
			if nextChar == "\"" {
				if i+3 < len(encoded) {
					nextAfterNext := encoded[i+2 : i+3]
					if nextAfterNext == "\"" {
						value = value + aChar + nextChar
						i += 2
						continue
					}
				}
				i++
				value = value + nextChar
				continue
			}
		}
		//allow unescaped " be inside text if the whole text is not enclosed in "s
		if aChar == "\"" && (len(value) == 0 || isInDoubleQuote) {
			isInDoubleQuote = !isInDoubleQuote
			continue
		}

		if encoded[i:i+1] == delimiter && !isInDoubleQuote {
			if !hasColumns {
				delimitedRecord.Columns = append(delimitedRecord.Columns, strings.TrimSpace(value))
			} else {
				var columnName = delimitedRecord.Columns[index]
				delimitedRecord.Record[columnName] = value
			}

			value = ""
			index++
			continue
		}
		value = value + aChar
	}
	if len(value) > 0 {
		if !hasColumns {
			delimitedRecord.Columns = append(delimitedRecord.Columns, strings.TrimSpace(value))
		} else {
			if index >= len(delimitedRecord.Columns) {
				return fmt.Errorf("index %v out of bound: columns: %v, values:%v", index, delimitedRecord.Columns, encoded)
			}
			var columnName = delimitedRecord.Columns[index]
			delimitedRecord.Record[columnName] = value
		}
	}
	return nil
}

type delimiterDecoderFactory struct{}

func (f *delimiterDecoderFactory) Create(reader io.Reader) Decoder {
	return &delimiterDecoder{reader: reader}
}

//NewDelimiterDecoderFactory returns a new delimitered decoder factory.
func NewDelimiterDecoderFactory() DecoderFactory {
	return &delimiterDecoderFactory{}
}

type yamlDecoderFactory struct{}

func (e yamlDecoderFactory) Create(reader io.Reader) Decoder {
	return &yamlDecoder{reader}
}

type yamlDecoder struct {
	io.Reader
}

func (d *yamlDecoder) Decode(target interface{}) error {
	var data, err = ioutil.ReadAll(d.Reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %T %v", d.Reader, err)
	}
	return yaml.Unmarshal(data, target)
}

//NewYamlDecoderFactory create a new yaml decoder factory
func NewYamlDecoderFactory() DecoderFactory {
	return &yamlDecoderFactory{}
}
