package toolbox_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
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

	record := &toolbox.DelimiteredRecord{
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
