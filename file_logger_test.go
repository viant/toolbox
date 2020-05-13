package toolbox_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"
)


func TestConfigLogger(t *testing.T) {

	_, err := toolbox.NewFileLogger(toolbox.FileLoggerConfig{
		LogType:           "test",
		FileTemplate:      "/tmp/test[yyyy].log",
		QueueFlashCount:   5,
		MaxQueueSize:      100,
		FlushRequencyInMs: 250,
		//MaxIddleTimeInSec: 2,
	})
	assert.NotNil(t, err)

	_, err = toolbox.NewFileLogger(toolbox.FileLoggerConfig{
		LogType:         "test",
		FileTemplate:    "/tmp/test[yyyy].log",
		QueueFlashCount: 5,
		MaxQueueSize:    100,
		//FlushRequencyInMs: 250,
		MaxIddleTimeInSec: 2,
	})
	assert.NotNil(t, err)

	_, err = toolbox.NewFileLogger(toolbox.FileLoggerConfig{
		LogType:         "test",
		FileTemplate:    "/tmp/test[yyyy].log",
		QueueFlashCount: 5,
		//MaxQueueSize      :100,
		FlushRequencyInMs: 250,
		MaxIddleTimeInSec: 2,
	})
	assert.NotNil(t, err)

	_, err = toolbox.NewFileLogger(toolbox.FileLoggerConfig{
		LogType:      "test",
		FileTemplate: "/tmp/test[yyyy].log",
		//QueueFlashCount        :5,
		MaxQueueSize:      100,
		FlushRequencyInMs: 250,
		MaxIddleTimeInSec: 2,
	})
	assert.NotNil(t, err)

	_, err = toolbox.NewFileLogger(toolbox.FileLoggerConfig{
		LogType: "test",
		//FileTemplate      :"/tmp/test[yyyy].log",
		QueueFlashCount:   5,
		MaxQueueSize:      100,
		FlushRequencyInMs: 250,
		MaxIddleTimeInSec: 2,
	})
	assert.NotNil(t, err)

	_, err = toolbox.NewFileLogger(toolbox.FileLoggerConfig{
		//LogType           :"test",
		FileTemplate:      "/tmp/test[yyyy].log",
		QueueFlashCount:   5,
		MaxQueueSize:      100,
		FlushRequencyInMs: 250,
		MaxIddleTimeInSec: 2,
	})
	assert.NotNil(t, err)

}

func TestLogger(t *testing.T) {

	testFile := fmt.Sprintf("/tmp/test%v.log", time.Now().Year())
	toolbox.RemoveFileIfExist(testFile)
	defer toolbox.RemoveFileIfExist(testFile)

	logger, err := toolbox.NewFileLogger(toolbox.FileLoggerConfig{
		LogType:           "test",
		FileTemplate:      "/tmp/test[yyyy].log",
		QueueFlashCount:   4,
		MaxQueueSize:      100,
		FlushRequencyInMs: 600,
		MaxIddleTimeInSec: 1,
	})

	assert.Nil(t, err)

	for i := 0; i < 6; i++ {
		logger.Log(&toolbox.LogMessage{
			MessageType: "test",
			Message:     fmt.Sprintf("Abc%v", i),
		})
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(400 * time.Millisecond)
	if file, err := os.Open(testFile); err == nil {
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		assert.Nil(t, err)
		assert.Equal(t, "Abc0\nAbc1\nAbc2\nAbc3\nAbc4\n", string(content))
	}

	time.Sleep(1 * time.Second)

	if file, err := os.Open(testFile); err == nil {
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		assert.Nil(t, err)
		assert.Equal(t, "Abc0\nAbc1\nAbc2\nAbc3\nAbc4\nAbc5\n", string(content))
	}

	time.Sleep(1 * time.Second)

	if file, err := os.Open(testFile); err == nil {
		file.Close()
		os.Remove(testFile)
	}

	logger.Notify(syscall.SIGKILL)

}

//
//func TestFileLogger_Notify(t *testing.T) {
//
//	testFile := fmt.Sprintf("/tmp/test%v.log", time.Now().Year())
//	toolbox.RemoveFileIfExist(testFile)
//	defer toolbox.RemoveFileIfExist(testFile)
//
//	logger, err := toolbox.NewFileLogger(toolbox.FileLoggerConfig{
//		LogType:           "test",
//		FileTemplate:      "/tmp/test[yyyy].log",
//		QueueFlashCount:   40,
//		MaxQueueSize:      100,
//		FlushRequencyInMs: 1200000,
//		MaxIddleTimeInSec: 20,
//	})
//
//	assert.Nil(t, err)
//
//	for i := 0; i < 6; i++ {
//		logger.Log(&toolbox.LogMessage{
//			MessageType: "test",
//			Message:     fmt.Sprintf("Abc%v", i),
//		})
//	}
//	logger.Notify(syscall.SIGKILL)
//	time.Sleep(100 * time.Millisecond)
//	if file, err := os.Open(testFile); err == nil {
//		defer file.Close()
//		content, err := ioutil.ReadAll(file)
//		assert.Nil(t, err)
//		assert.Equal(t, "Abc0\nAbc1\nAbc2\nAbc3\nAbc4\nAbc5\n", string(content))
//	}
//	assert.Nil(t, err)
//}
