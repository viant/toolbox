package storage_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/storage"
	"os"
	"testing"
)

func Test_NewFileMode(t *testing.T) {

	var testData = map[string]int{
		"drwxr-xr-x": 0x800001ed,
		"drwxrwxrwx": 0x800001ff,
		"drwxr-----": 0x800001e0,
		"prw-rw-rw-": 0x20001b6,
	}
	for attr, mode := range testData {
		var attributeMode, err = storage.NewFileMode(attr)
		assert.Nil(t, err)
		var fileMode = os.FileMode(mode)
		assert.Equal(t, int(fileMode), int(attributeMode))
	}

}
