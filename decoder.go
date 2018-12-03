package toolbox

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
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
		return fmt.Errorf("failed to decode - unable cast %T to %s", v, result)
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
		return fmt.Errorf("invalid target type, expected %T but had %T", &DelimitedRecord{}, target)
	}
	if delimitedRecord.Record == nil {
		delimitedRecord.Record = make(map[string]interface{})
	}
	var delimiter = delimitedRecord.Delimiter

	payload, err := ioutil.ReadAll(d.reader)
	if err != nil {
		return err
	}
	reader := csv.NewReader(bytes.NewReader(payload))
	reader.Comma = rune(delimiter[0])
	hasColumns := len(delimitedRecord.Columns) > 0
	if !hasColumns {
		delimitedRecord.Columns = make([]string, 0)
	}
	record, err := reader.Read()
	if IsEOFError(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(delimitedRecord.Columns) == 0 {
		for _, field := range record {
			delimitedRecord.Columns = append(delimitedRecord.Columns, strings.TrimSpace(field))
		}
	} else {
		for i, field := range record {
			delimitedRecord.Record[delimitedRecord.Columns[i]] = field
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

type flexYamlDecoderFactory struct{}

func (e flexYamlDecoderFactory) Create(reader io.Reader) Decoder {
	return &flexYamlDecoder{reader}
}

type flexYamlDecoder struct {
	io.Reader
}

//normalizeMap normalizes keyValuePairs from map or slice (map with preserved key order)
func (d *flexYamlDecoder) normalizeMap(keyValuePairs interface{}, deep bool) (map[string]interface{}, error) {
	var result = make(map[string]interface{})
	if keyValuePairs == nil {
		return result, nil
	}
	err := ProcessMap(keyValuePairs, func(k, value interface{}) bool {
		var key = AsString(k)

		//inline map key
		result[key] = value
		if deep {
			if value == nil {
				return true
			}
			if IsMap(value) {
				if normalized, err := d.normalizeMap(value, deep); err == nil {
					result[key] = normalized
				}
			} else if IsSlice(value) { //yaml style map conversion if applicable
				aSlice := AsSlice(value)
				if len(aSlice) == 0 {
					return true
				}
				if IsMap(aSlice[0]) || IsStruct(aSlice[0]) {
					normalized, err := d.normalizeMap(value, deep)
					if err == nil {
						result[key] = normalized
					}
				} else if IsSlice(aSlice[0]) {
					for i, item := range aSlice {
						itemMap, err := d.normalizeMap(item, deep)
						if err != nil {
							return true
						}
						aSlice[i] = itemMap
					}
					result[key] = aSlice
				}
				return true
			}
		}
		return true
	})
	return result, err
}

func (d *flexYamlDecoder) Decode(target interface{}) error {
	var data, err = ioutil.ReadAll(d.Reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %T %v", d.Reader, err)
	}
	aMap := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &aMap); err != nil {
		return err
	}
	if normalized, err := d.normalizeMap(aMap, true); err == nil {
		aMap = normalized
	}
	return DefaultConverter.AssignConverted(target, aMap)
}

//NewFlexYamlDecoderFactory create a new yaml decoder factory
func NewFlexYamlDecoderFactory() DecoderFactory {
	return &flexYamlDecoderFactory{}
}
