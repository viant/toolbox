package data

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ExtractPath(t *testing.T) {

	assert.EqualValues(t, "var", ExtractPath("++$var"))
	assert.EqualValues(t, "var.z", ExtractPath("->${var.z}"))

	assert.EqualValues(t, "var.key[0]", ExtractPath("<-$var.key[0]"))
	assert.EqualValues(t, "var.z", ExtractPath("->${var.z}"))

}

