package toolbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

//IsStructuredJSON returns true if supplied represent JSON structure (map,array)
func IsStructuredJSON(candidate string) bool {
	candidate = strings.Trim(candidate, "\n \t\r")
	if candidate == "" {
		return false
	}

	curlyStart := strings.Count(candidate, "{")
	curlyEnd := strings.Count(candidate, "}")

	squareStart := strings.Count(candidate, "[")
	squareEnd := strings.Count(candidate, "]")
	if !(curlyStart == curlyEnd && squareStart == squareEnd) || (curlyStart+squareStart == 0) {
		return false
	}
	if !(strings.HasPrefix(candidate, "{") && strings.HasSuffix(candidate, "}") || strings.HasPrefix(candidate, "[") && strings.HasSuffix(candidate, "]")) {
		return false
	}
	return json.Valid([]byte(candidate))
}

//IsCompleteJSON returns true if supplied represent complete JSON
func IsCompleteJSON(candidate string) bool {
	return json.Valid([]byte(candidate))
}

//NewLineDelimitedJSON returns JSON for supplied multi line JSON
func NewLineDelimitedJSON(candidate string) ([]interface{}, error) {
	var result = make([]interface{}, 0)
	lines := getMultilineContent(candidate)
	for _, line := range lines {
		aStruct, err := JSONToInterface(line)
		if err != nil {
			return nil, err
		}
		result = append(result, aStruct)
	}
	return result, nil
}

func getMultilineContent(multiLineText string) []string {
	multiLineText = strings.TrimSpace(multiLineText)
	if multiLineText == "" {
		return []string{}
	}
	lines := strings.Split(multiLineText, "\n")
	var result = make([]string, 0)
	for _, line := range lines {
		if strings.Trim(line, " \r") == "" {
			continue
		}
		result = append(result, line)
	}
	return result
}

//IsNewLineDelimitedJSON returns true if supplied content is multi line delimited JSON
func IsNewLineDelimitedJSON(candidate string) bool {
	lines := getMultilineContent(candidate)
	if len(lines) <= 1 {
		return false
	}
	return IsStructuredJSON(lines[0]) && IsStructuredJSON(lines[1])
}

//JSONToInterface converts JSON source to an interface (either map or slice)
func JSONToInterface(source interface{}) (interface{}, error) {
	var reader io.Reader
	switch value := source.(type) {
	case io.Reader:
		reader = value
	case []byte:
		reader = bytes.NewReader(value)
	case string:
		reader = strings.NewReader(value)
	default:
		return nil, fmt.Errorf("unsupported type: %T", source)
	}
	var result interface{}
	if content, err := ioutil.ReadAll(reader); err == nil {
		text := string(content)
		if IsNewLineDelimitedJSON(text) {
			return NewLineDelimitedJSON(text)
		}
		reader = strings.NewReader(text)
	}
	err := jsonDecoderFactory{}.Create(reader).Decode(&result)
	return result, err
}

//JSONToMap converts JSON source into map
func JSONToMap(source interface{}) (map[string]interface{}, error) {
	var reader io.Reader
	switch value := source.(type) {
	case io.Reader:
		reader = value
	case []byte:
		reader = bytes.NewReader(value)
	case string:
		reader = strings.NewReader(value)
	default:
		return nil, fmt.Errorf("unsupported type: %T", source)
	}
	var result = make(map[string]interface{})
	err := jsonDecoderFactory{}.Create(reader).Decode(&result)
	return result, err
}

//JSONToSlice converts JSON source into slice
func JSONToSlice(source interface{}) ([]interface{}, error) {
	var reader io.Reader
	switch value := source.(type) {
	case io.Reader:
		reader = value
	case []byte:
		reader = bytes.NewReader(value)
	case string:
		reader = strings.NewReader(value)
	default:
		return nil, fmt.Errorf("unsupported type: %T", source)
	}
	var result = make([]interface{}, 0)
	err := jsonDecoderFactory{}.Create(reader).Decode(&result)
	return result, err
}

//AsJSONText converts data structure int text JSON
func AsJSONText(source interface{}) (string, error) {
	if source == nil {
		return "", fmt.Errorf("source was nil")
	}
	if IsStruct(source) || IsMap(source) || IsSlice(source) {
		buf := new(bytes.Buffer)
		err := NewJSONEncoderFactory().Create(buf).Encode(source)
		return buf.String(), err
	}
	return "", fmt.Errorf("unsupported type: %T", source)
}

//AsIndentJSONText converts data structure int text JSON
func AsIndentJSONText(source interface{}) (string, error) {
	if IsStruct(source) || IsMap(source) || IsSlice(source) {
		buf, err := json.MarshalIndent(source, "", "\t")
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	return "", fmt.Errorf("unsupported type: %T", source)
}

//AnyJSONType represents any JSON type
type AnyJSONType string

//UnmarshalJSON implements unmarshalerinterface
func (s *AnyJSONType) UnmarshalJSON(b []byte) error {
	*s = AnyJSONType(b)
	return nil
}

//MarshalJSON implements marshaler interface
func (s *AnyJSONType) MarshalJSON() ([]byte, error) {
	if len(*s) == 0 {
		return []byte(`""`), nil
	}
	return []byte(*s), nil
}

//Value returns string or string slice value
func (s AnyJSONType) Value() (interface{}, error) {
	var result interface{}
	return result, json.Unmarshal([]byte(s), &result)
}
