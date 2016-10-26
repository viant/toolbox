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
