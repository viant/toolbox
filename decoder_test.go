package toolbox_test

import (
	"testing"
	"github.com/viant/toolbox"
	"strings"
	"github.com/stretchr/testify/assert"
)

func TestDecoderFactory(t *testing.T) {
	reader := strings.NewReader("[1, 2, 3]")
	decoder := toolbox.NewJSONDecoderFactory().Create(reader)
	aSlice := make([]int,0)
	decoder.Decode(&aSlice)
	assert.Equal(t, 3, len(aSlice))
}
