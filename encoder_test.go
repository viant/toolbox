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
