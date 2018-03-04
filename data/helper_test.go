package data

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_ExtractPath(t *testing.T) {

	assert.EqualValues(t, "var", ExtractPath("++$var"))
	assert.EqualValues(t, "var.z", ExtractPath("->${var.z}"))
	assert.EqualValues(t, "var.key[0]", ExtractPath("<-$var.key[0]"))
	assert.EqualValues(t, "var.z", ExtractPath("->${var.z}"))

}


