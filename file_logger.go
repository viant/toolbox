package toolbox

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

//FileLoggerConfig represents FileLogger
type FileLoggerConfig struct {
	LogType           string
	FileTemplate      string
	QueueFlashCount   int
	MaxQueueSize      int
	FlushRequencyInMs int
	MaxIddleTimeInSec int
}

//Validate valides configuration sttings
func (c *FileLoggerConfig) Validate() error {
	if len(c.LogType) == 0 {
		return errors.New("Log type was empty")
	}
	if c.FlushRequencyInMs == 0 {
		return errors.New("FlushRequencyInMs was 0")
	}
	if c.MaxQueueSize == 0 {
		return errors.New("MaxQueueSize was 0")
	}
	if len(c.FileTemplate) == 0 {
		return errors.New("FileTemplate was empty")
	}
	if c.MaxIddleTimeInSec == 0 {
		return errors.New("MaxIddleTimeInSec was 0")
	}
	if c.QueueFlashCount == 0 {
		return errors.New("QueueFlashCount was 0")
	}
	return nil
}

//LogStream represents individual log stream
type LogStream struct {
	Name             string
	Logger           *FileLogger
	Config           *FileLoggerConfig
	RecordCount      int
	File             *os.File
	LastAddQueueTime time.Time
	LastWriteTime    uint64
	Messages         chan string
	Complete         chan bool
}

//Log logs message into stream
func (s *LogStream) Log(message *LogMessage) {
	textMessage := message.Message.(string)
	s.Messages <- textMessage
	s.LastAddQueueTime = time.Now()
}

func (s *LogStream) write(message string) error {
	atomic.StoreUint64(&s.LastWriteTime, uint64(time.Now().UnixNano()))
	_, err := s.File.WriteString(message)
	if err != nil {
		return err
	}
	return s.File.Sync()
}

//Close closes stream.
func (s *LogStream) Close() {
	s.Logger.streamMapMutex.Lock()
	delete(s.Logger.streams, s.Name)
	s.Logger.streamMapMutex.Unlock()
	s.File.Close()

}

func (s *LogStream) isFrequencyFlushNeeded() bool {
	elapsedInMs := (int(time.Now().UnixNano()) - int(atomic.LoadUint64(&s.LastWriteTime))) / 1000000
	return elapsedInMs >= s.Config.FlushRequencyInMs
}

func (s *LogStream) manageWritesInBatch() {
	messageCount := 0
	var message, messages string
	var timeout = time.Duration(2 * int(s.Config.FlushRequencyInMs) * int(time.Millisecond))
	for {
		select {
		case done := <-s.Complete:
			if done {
				manageWritesInBatchLoopFlush(s, messageCount, messages)
				s.Close()
				os.Exit(0)
			}
		case <-time.After(timeout):
			if !manageWritesInBatchLoopFlush(s, messageCount, messages) {
				return
			} else {
				messageCount = 0
				messages = ""
			}
		case message = <-s.Messages:
			messages += message + "\n"

			messageCount++
			s.RecordCount++

			var hasReachMaxRecrods = messageCount >= s.Config.QueueFlashCount && s.Config.QueueFlashCount > 0
			if hasReachMaxRecrods || s.isFrequencyFlushNeeded() {
				s.write(messages)
				messages = ""
				messageCount = 0
			}

		}
	}
}

func manageWritesInBatchLoopFlush(s *LogStream, messageCount int, messages string) bool {
	if messageCount > 0 {
		if s.isFrequencyFlushNeeded() {
			err := s.write(messages)
			if err != nil {
				fmt.Printf("Failed to write to log due to %v", err)
			}
			return true
		}
	}
	elapsedInMs := (int(time.Now().UnixNano()) - int(atomic.LoadUint64(&s.LastWriteTime))) / 1000000
	if elapsedInMs > s.Config.MaxIddleTimeInSec*1000 {
		s.Close()
		return false
	}
	return true
}

//FileLogger represents a file logger
type FileLogger struct {
	config         map[string]*FileLoggerConfig
	streamMapMutex *sync.Mutex
	streams        map[string]*LogStream
}

func (l *FileLogger) getConfig(messageType string) (*FileLoggerConfig, error) {
	config, found := l.config[messageType]
	if !found {
		return nil, errors.New("Failed to lookup config for " + messageType)
	}
	return config, nil
}

//ExpandFileTemplate expands
func ExpandFileTemplate(template string) string {
	startIndex := strings.Index(template, "[")
	if startIndex == -1 {
		return template
	}
	endIndex := strings.Index(template, "]")
	if endIndex == -1 {
		return template
	}
	format := template[startIndex+1 : endIndex]

	formatedTime := time.Now().Format(DateFormatToLayout(format))
	source := "[" + format + "]"
	return strings.Replace(template, source, formatedTime, 1)
}

//NewLogStream creat a new LogStream for passed om path and file config
func (l *FileLogger) NewLogStream(path string, config *FileLoggerConfig) (*LogStream, error) {
	osFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	logStream := &LogStream{Name: path, Logger: l, Config: config, File: osFile, Messages: make(chan string, config.MaxQueueSize), Complete: make(chan bool)}
	go func() {
		logStream.manageWritesInBatch()
	}()
	return logStream, nil
}

func (l *FileLogger) acquireLogStream(messageType string) (*LogStream, error) {
	config, err := l.getConfig(messageType)
	if err != nil {
		return nil, err
	}
	fileName := ExpandFileTemplate(config.FileTemplate)
	l.streamMapMutex.Lock()
	defer l.streamMapMutex.Unlock()
	logStream, found := l.streams[fileName]

	if found {
		return logStream, nil
	}
	logStream, err = l.NewLogStream(fileName, config)
	if err != nil {
		return nil, err
	}
	l.streams[fileName] = logStream
	return logStream, nil
}

//Log logs message into stream
func (l *FileLogger) Log(message *LogMessage) error {
	logStream, err := l.acquireLogStream(message.MessageType)
	if err != nil {
		return err
	}
	logStream.Log(message)
	return nil
}

//NewFileLogger create new file logger
func NewFileLogger(configs ...FileLoggerConfig) (*FileLogger, error) {
	result := &FileLogger{
		config:         make(map[string]*FileLoggerConfig),
		streamMapMutex: &sync.Mutex{},
		streams:        make(map[string]*LogStream),
	}

	for i := range configs {
		err := configs[i].Validate()
		if err != nil {
			return nil, err
		}
		result.config[configs[i].LogType] = &configs[i]
	}
	// If there's a signal to quit the program send it to channel
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)
	go func() {
		// Block until receive a quit signal
		_quit := <-sigc
		_ = _quit // don't care which type
		for _, value := range result.streams {
			// No wait flush
			value.Config.FlushRequencyInMs = 0
			// Write logs now
			value.Complete <- true
		}
	}()

	return result, nil
}
