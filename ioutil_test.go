package toolbox_test

import (
	"testing"
	"strings"
	"runtime"
	"github.com/viant/toolbox"
	"os"
	"github.com/stretchr/testify/assert"
)


func TestOpenURL(t *testing.T) {
	fileName, _, _ := getCallerInfo(2)
	{
		file, err := toolbox.OpenURL(toolbox.FileSchema + fileName, os.O_RDONLY, 0644)
		assert.Nil(t, err)
		defer file.Close()
	}
	{
		_, err := toolbox.OpenURL(toolbox.FileSchema + fileName + "bleh_bleh", os.O_RDONLY, 0644)
		assert.NotNil(t, err)
	}

	{
		_, err := toolbox.OpenURL("https://github.com/viant/toolbox", os.O_RDONLY, 0644)
		assert.NotNil(t, err, "only file protocol is supported")
	}



}

func TestOpenReaderFromURL(t *testing.T) {
	fileName, _, _ := getCallerInfo(2)
	{
		file, _, err := toolbox.OpenReaderFromURL(toolbox.FileSchema + fileName)
		assert.Nil(t, err)
		defer file.Close()
	}
	{
		_, _, err := toolbox.OpenReaderFromURL(toolbox.FileSchema + fileName + "blahbla")
		assert.NotNil(t, err)
	}

	{
		file, _, err := toolbox.OpenReaderFromURL("https://github.com/viant/toolbox")
		assert.Nil(t, err)
		defer file.Close()
	}

	{
		_, _, err := toolbox.OpenReaderFromURL("abc://github.com/viant/toolbox")
		assert.NotNil(t, err)
	}
}


func getCallerInfo(callerIndex int) (string, string, int) {
	var callerPointer = make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(callerIndex, callerPointer)
	callerInfo := runtime.FuncForPC(callerPointer[0])
	file, line := callerInfo.FileLine(callerPointer[0])
	callerName := callerInfo.Name()
	dotPosition := strings.LastIndex(callerName, ".")
	return file, callerName[dotPosition+1 : len(callerName)], line
}