package toolbox_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestDecoderFactory(t *testing.T) {
	reader := strings.NewReader("[1, 2, 3]")
	decoder := toolbox.NewJSONDecoderFactory().Create(reader)
	aSlice := make([]int, 0)
	err := decoder.Decode(&aSlice)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(aSlice))
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
