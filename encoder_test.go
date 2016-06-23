package toolbox_test

import (
	"github.com/viant/toolbox"
	"github.com/stretchr/testify/assert"
	"testing"
	"bytes"
)

func TestEncoderFactory(t *testing.T) {
	buffer := new(bytes.Buffer)
	assert.NotNil(t, toolbox.NewJSONEncoderFactory().Create(buffer))
}




