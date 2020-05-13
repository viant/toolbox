package toolbox

import (
	"bytes"
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
	LogType            string
	FileTemplate       string
	filenameProvider   func(t time.Time) string
	QueueFlashCount    int
	MaxQueueSize       int
	FlushRequencyInMs  int //type backward-forward compatibility
	FlushFrequencyInMs int
	MaxIddleTimeInSec  int
	inited             bool
}

func (c *FileLoggerConfig) Init() {
	if c.inited {
		return
	}
	defaultProvider := func(t time.Time) string {
		return c.FileTemplate
	}
	c.inited = true
	template := c.FileTemplate
	c.filenameProvider = defaultProvider
	startIndex := strings.Index(template, "[")
	if startIndex == -1 {
		return
	}
	endIndex := strings.Index(template, "]")
	if endIndex == -1 {
		return
	}
	format := template[startIndex+1 : endIndex]
	layout := DateFormatToLayout(format)
	c.filenameProvider = func(t time.Time) string {
		formatted := t.Format(layout)
		return strings.Replace(template, "["+format+"]", formatted, 1)
	}
}

//Validate valides configuration sttings
func (c *FileLoggerConfig) Validate() error {
	if len(c.LogType) == 0 {
		return errors.New("Log type was empty")
	}
	if c.FlushFrequencyInMs == 0 {
		c.FlushFrequencyInMs = c.FlushRequencyInMs
	}
	if c.FlushFrequencyInMs == 0 {
		return errors.New("FlushFrequencyInMs was 0")
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
func (s *LogStream) Log(message *LogMessage) error {
	if message == nil {
		return errors.New("message was nil")
	}
	var textMessage = ""
	var ok bool
	if textMessage, ok = message.Message.(string); ok {
	} else if IsStruct(message.Message) || IsMap(message.Message) || IsSlice(message.Message) {
		var buf = new(bytes.Buffer)
		err := NewJSONEncoderFactory().Create(buf).Encode(message.Message)
		if err != nil {
			return err
		}
		textMessage = strings.Trim(buf.String(), "\n\r")
	} else {
		return fmt.Errorf("unsupported type: %T", message.Message)
	}
	s.Messages <- textMessage
	s.LastAddQueueTime = time.Now()
	return nil
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
	return elapsedInMs >= s.Config.FlushFrequencyInMs
}

func (s *LogStream) manageWritesInBatch() {
	messageCount := 0
	var message, messages string
	var timeout = time.Duration(2 * int(s.Config.FlushFrequencyInMs) * int(time.Millisecond))
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
			}
			messageCount = 0
			messages = ""
		case message = <-s.Messages:
			messages += message + "\n"
			messageCount++
			s.RecordCount++

			var hasReachMaxRecrods = messageCount >= s.Config.QueueFlashCount && s.Config.QueueFlashCount > 0
			if hasReachMaxRecrods || s.isFrequencyFlushNeeded() {
				_ = s.write(messages)
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
				fmt.Printf("failed to write to log due to %v", err)
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
	streamMapMutex *sync.RWMutex
	streams        map[string]*LogStream
	siginal        chan os.Signal
}

func (l *FileLogger) getConfig(messageType string) (*FileLoggerConfig, error) {
	config, found := l.config[messageType]
	if !found {
		return nil, errors.New("failed to lookup config for " + messageType)
	}
	config.Init()
	return config, nil
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
	fileName := config.filenameProvider(time.Now())
	l.streamMapMutex.RLock()
	logStream, found := l.streams[fileName]
	l.streamMapMutex.RUnlock()
	if found {
		return logStream, nil
	}

	logStream, err = l.NewLogStream(fileName, config)
	if err != nil {
		return nil, err
	}
	l.streamMapMutex.Lock()
	l.streams[fileName] = logStream
	l.streamMapMutex.Unlock()
	return logStream, nil
}

//Log logs message into stream
func (l *FileLogger) Log(message *LogMessage) error {
	logStream, err := l.acquireLogStream(message.MessageType)
	if err != nil {
		return err
	}
	return logStream.Log(message)
}

//Notify notifies logger
func (l *FileLogger) Notify(siginal os.Signal) {
	l.siginal <- siginal
}

//NewFileLogger create new file logger
func NewFileLogger(configs ...FileLoggerConfig) (*FileLogger, error) {
	result := &FileLogger{
		config:         make(map[string]*FileLoggerConfig),
		streamMapMutex: &sync.RWMutex{},
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
	result.siginal = make(chan os.Signal, 1)
	signal.Notify(result.siginal,
		syscall.SIGINT,
		syscall.SIGTERM)

	go func() {

		// Block until receive a quit signal
		_quit := <-result.siginal
		_ = _quit // don't care which type
		for _, stream := range result.streams {
			// No wait flush
			stream.Config.FlushFrequencyInMs = 0
			// Write logs now
			stream.Complete <- true
		}
	}()

	return result, nil
}
