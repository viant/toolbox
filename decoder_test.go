package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"os"
	"path"
	"strings"
	"testing"
)

func TestDecoderFactory(t *testing.T) {
	{
		reader := strings.NewReader("[1, 2, 3]")
		decoder := toolbox.NewJSONDecoderFactory().Create(reader)
		aSlice := make([]int, 0)
		err := decoder.Decode(&aSlice)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(aSlice))
	}
	{
		reader := strings.NewReader("[1, 2, 3]")
		decoder := toolbox.NewJSONDecoderFactoryWithOption(true).Create(reader)
		aSlice := make([]int, 0)
		err := decoder.Decode(&aSlice)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(aSlice))
	}
}

func TestUnMarshalerDecoderFactory(t *testing.T) {
	reader := strings.NewReader("abc")
	decoder := toolbox.NewUnMarshalerDecoderFactory().Create(reader)
	foo := &Foo100{}
	err := decoder.Decode(foo)
	assert.Nil(t, err)
	assert.Equal(t, "abc", foo.Attr)

	err = decoder.Decode(&Foo101{})
	assert.NotNil(t, err)

}

type Foo100 struct {
	Attr string
}

func (m *Foo100) Unmarshal(data []byte) error {
	m.Attr = string(data)
	return nil
}

type Foo101 struct {
	Attr string
}

func TestDelimiterDecoderFactory(t *testing.T) {

	record := &toolbox.DelimitedRecord{
		Delimiter: ",",
	}
	{
		decoder := toolbox.NewDelimiterDecoderFactory().Create(strings.NewReader("column1,\"column2\", column3,column4"))
		err := decoder.Decode(record)
		if assert.Nil(t, err) {
			assert.Equal(t, []string{"column1", "column2", "column3", "column4"}, record.Columns)

		}
	}

	{
		decoder := toolbox.NewDelimiterDecoderFactory().Create(strings.NewReader("1,2,\"ab,cd\",3"))
		err := decoder.Decode(record)
		if assert.Nil(t, err) {
			assert.EqualValues(t, "1", record.Record["column1"])
			assert.EqualValues(t, "2", record.Record["column2"])
			assert.EqualValues(t, "ab,cd", record.Record["column3"])
			assert.EqualValues(t, "3", record.Record["column4"])
		}
	}

	{
		decoder := toolbox.NewDelimiterDecoderFactory().Create(strings.NewReader("1,2,\" \"\"location:[\\\"\"BE\\\"\"]\"\"  \",3"))
		err := decoder.Decode(record)
		if assert.Nil(t, err) {
			assert.EqualValues(t, "1", record.Record["column1"])
			assert.EqualValues(t, "2", record.Record["column2"])
			assert.EqualValues(t, " \"location:[\\\"BE\\\"]\"  ", record.Record["column3"])
			assert.EqualValues(t, "3", record.Record["column4"])
		}
	}

}

func TestTestYamlDecoder(t *testing.T) {
	var filename = path.Join(os.Getenv("TMPDIR"), "test.yaml")
	toolbox.RemoveFileIfExist(filename)
	defer toolbox.RemoveFileIfExist(filename)
	var aMap = map[string]interface{}{
		"a": 1,
		"b": "123",
		"c": []int{1, 3, 6},
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if assert.Nil(t, err) {
		err = toolbox.NewYamlEncoderFactory().Create(file).Encode(aMap)
		assert.Nil(t, err)
	}
	var cloneMap = make(map[string]interface{})
	file.Close()
	file, err = os.OpenFile(filename, os.O_RDONLY, 0644)
	if assert.Nil(t, err) {
		defer file.Close()
		err = toolbox.NewYamlDecoderFactory().Create(file).Decode(&cloneMap)
		if assert.Nil(t, err) {
			assert.EqualValues(t, aMap["a"], cloneMap["a"])
			assert.EqualValues(t, aMap["b"], cloneMap["b"])
			assert.EqualValues(t, toolbox.AsSlice(aMap["c"]), cloneMap["c"])

		}
	}

}
