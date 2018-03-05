package toolbox_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestEncoderFactory(t *testing.T) {
	buffer := new(bytes.Buffer)
	assert.NotNil(t, toolbox.NewJSONEncoderFactory().Create(buffer))
}

func TestMarshalEncoderFactory(t *testing.T) {
	buffer := new(bytes.Buffer)
	encoder := toolbox.NewMarshalerEncoderFactory().Create(buffer)
	foo := &Foo200{"abc"}
	err := encoder.Encode(foo)
	assert.Nil(t, err)
	assert.Equal(t, "abc", string(buffer.Bytes()))
	err = encoder.Encode(&Foo201{})
	assert.NotNil(t, err)
}

type Foo200 struct {
	Attr string
}

func (m *Foo200) Marshal() ([]byte, error) {
	return []byte(m.Attr), nil
}

type Foo201 struct {
	Attr string
}
