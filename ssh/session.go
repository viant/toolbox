package ssh

import (
	"bytes"
	"fmt"
	"github.com/lunixbochs/vtclean"
	"github.com/pkg/errors"
	"github.com/viant/toolbox"
	"golang.org/x/crypto/ssh"
	"io"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type TerminatedError struct {
	Err error
}

func (t *TerminatedError) Error() string {
	return fmt.Sprintf("terminated due to %v", t.Err)
}

//ErrTerminated - command session terminated
var ErrTerminated = &TerminatedError{}

const defaultShell = "/bin/bash"

const (
	drainTimeoutMs         = 10
	defaultTimeoutMs       = 20000
	stdoutFlashFrequencyMs = 1000
	initTimeoutMs          = 300
	defaultTickFrequency   = 100
)

//Listener represent command listener (it will send stdout fragments as thier being available on stdout)
type Listener func(stdout string, hasMore bool)

//MultiCommandSession represents a multi command session
type MultiCommandSession interface {
	Run(command string, listener Listener, timeoutMs int, terminators ...string) (string, error)

	ShellPrompt() string

	System() string

	Reconnect() error

	Close()
}

//multiCommandSession represents a multi command session
//a new command are send vi stdin
type multiCommandSession struct {
	service            *service
	config             *SessionConfig
	replayCommands     *ReplayCommands
	recordSession      bool
	session            *ssh.Session
	stdOutput          chan string
	stdError           chan string
	stdInput           io.WriteCloser
	promptSequence     string
	shellPrompt        string
	escapedShellPrompt string
	system             string
	running            int32
	stdin              string
}

func (s *multiCommandSession) Run(command string, listener Listener, timeoutMs int, terminators ...string) (string, error) {
	if atomic.LoadInt32(&s.running) == 0 {
		return "", ErrTerminated
	}
	s.drainStdout()
	if !strings.HasSuffix(command, "\n") {
		command += "\n"
	}
	var stdin = command
	s.stdin = stdin
	_, err := s.stdInput.Write([]byte(stdin))
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v, err: %v", command, err)
	}
	var output string
	output, _, err = s.readResponse(timeoutMs, listener, terminators...)
	if s.recordSession {
		s.replayCommands.Register(stdin, output)
	}
	return output, err
}

//ShellPrompt returns a shell prompt
func (s *multiCommandSession) ShellPrompt() string {
	return s.shellPrompt
}

//System returns a system name
func (s *multiCommandSession) System() string {
	return s.system
}

//Close closes the session with its resources
func (s *multiCommandSession) Close() {
	atomic.StoreInt32(&s.running, 0)
	_ = s.stdInput.Close()
	if s.session != nil {
		_ = s.session.Close()
	}

}

func (s *multiCommandSession) closeIfError(err error) bool {
	if err != nil {
		if err != ErrTerminated {
			ErrTerminated.Err = err
		}
		s.Close()
		return true
	}
	return false
}

func (s *multiCommandSession) start(shell string) (output string, err error) {
	var reader, errReader io.Reader
	reader, err = s.session.StdoutPipe()
	if err != nil {
		return "", err
	}
	errReader, err = s.session.StderrPipe()
	if err != nil {
		return "", err
	}

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)
	go func() {
		waitGroup.Done()
		s.copy(reader, s.stdOutput)
	}()
	go func() {
		waitGroup.Done()
		s.copy(errReader, s.stdError)
	}()

	if shell == "" {
		shell = defaultShell
	}
	waitGroup.Wait()
	s.stdin = shell
	err = s.session.Start(shell)
	if err != nil {
		return "", err
	}
	_, name := path.Split(shell)
	output, _, err = s.readResponse(defaultTimeoutMs, nil, name)
	return output, err
}

//copy copy data from reader to channel
func (s *multiCommandSession) copy(reader io.Reader, out chan string) {
	var written int64 = 0
	buf := make([]byte, 128*1024)
	var err error
	var bytesRead int
	for {
		writer := new(bytes.Buffer)
		if atomic.LoadInt32(&s.running) == 0 {
			return
		}
		bytesRead, err = reader.Read(buf)
		if bytesRead > 0 {
			bytesWritten, writeError := writer.Write(buf[:bytesRead])
			if s.closeIfError(writeError) {
				return
			}
			if bytesWritten > 0 {
				written += int64(bytesWritten)
			}

			if bytesRead != bytesWritten {
				if s.closeIfError(io.ErrShortWrite) {
					return
				}
			}
			out <- string(writer.Bytes())
		}

		if s.closeIfError(err) {
			return
		}
	}
}

func escapeInput(input string) string {
	input = vtclean.Clean(input, false)
	if input == "" {
		return input
	}
	return strings.Trim(input, "\n\r\t ")
}

func (s *multiCommandSession) Reconnect() (err error) {
	atomic.StoreInt32(&s.running, 1)
	s.service.Reconnect()
	s.session, err = s.service.client.NewSession()
	defer func() {
		if err != nil {
			s.service.client.Close()
		}
	}()
	if err != nil {
		return err
	}
	return s.init()
}

func (s *multiCommandSession) hasPrompt(input string) bool {
	escapedInput := escapeInput(input)
	var shellPrompt = s.shellPrompt
	if shellPrompt == "" {
		shellPrompt = "$"
	}
	if s.escapedShellPrompt == "" && s.shellPrompt != "" {
		s.escapedShellPrompt = escapeInput(s.shellPrompt)
	}

	if s.escapedShellPrompt != "" && strings.HasSuffix(escapedInput, s.escapedShellPrompt) || strings.HasSuffix(input, s.shellPrompt) {
		return true
	}
	return false
}

func (s *multiCommandSession) hasTerminator(input string, terminators ...string) bool {
	escapedInput := escapeInput(input)
	input = escapedInput
	for _, candidate := range terminators {
		candidateLen := len(candidate)
		if candidateLen == 0 {
			continue
		}
		if candidate[0:1] == "^" && strings.HasPrefix(input, candidate[1:]) {
			return true
		}
		if candidate[candidateLen-1:] == "$" && strings.HasSuffix(input, candidate[:candidateLen-1]) {
			return true
		}
		if strings.Contains(input, candidate) {
			return true
		}
	}
	return false
}

func (s *multiCommandSession) removePromptIfNeeded(stdout string) string {
	if strings.Contains(stdout, s.shellPrompt) {
		stdout = strings.Replace(stdout, s.shellPrompt, "", 1)
		var lines = []string{}
		for _, line := range strings.Split(stdout, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			lines = append(lines, line)
		}
		stdout = strings.Join(lines, "\n")
	}
	return stdout
}

func (s *multiCommandSession) readResponse(timeoutMs int, listener Listener, terminators ...string) (out string, has bool, err error) {
	var hasPrompt, hasTerminator bool
	if timeoutMs == 0 {
		timeoutMs = defaultTimeoutMs
	}
	notification := newNotificationWindow(listener, stdoutFlashFrequencyMs)
	defer notification.flush()

	var done int32
	defer atomic.StoreInt32(&done, 1)
	var errOut string
	var hasOutput bool

	var waitTimeMs = 0
	var tickFrequencyMs = defaultTickFrequency
	if tickFrequencyMs > timeoutMs {
		tickFrequencyMs = timeoutMs
	}
	var timeoutDuration = time.Duration(tickFrequencyMs) * time.Millisecond

outer:
	for {
		select {
		case partialOutput := <-s.stdOutput:
			waitTimeMs = 0
			out += partialOutput
			hasTerminator = s.hasTerminator(out, terminators...)
			if len(partialOutput) > 0 {
				if hasTerminator {
					partialOutput = addLineBreakIfNeeded(partialOutput)
				}
				notification.notify(s.removePromptIfNeeded(partialOutput))
			}
			hasPrompt = s.hasPrompt(out)
			if (hasPrompt || hasTerminator) && len(s.stdOutput) == 0 {
				break outer
			}
		case e := <-s.stdError:
			errOut += e
			notification.notify(s.removePromptIfNeeded(e))
			hasPrompt = s.hasPrompt(errOut)
			hasTerminator = s.hasTerminator(errOut, terminators...)
			if (hasPrompt || hasTerminator) && len(s.stdOutput) == 0 {
				break outer
			}
		case <-time.After(timeoutDuration):
			waitTimeMs += tickFrequencyMs
			if waitTimeMs >= timeoutMs {
				break outer
			}
		}
	}

	if hasTerminator {
		s.drainStdout()

	}
	if errOut != "" {
		err = errors.New(errOut)
	}

	if len(out) > 0 {
		hasOutput = true
		out = s.removePromptIfNeeded(out)
	}
	return out, hasOutput, err
}
func addLineBreakIfNeeded(text string) string {
	index := strings.LastIndex(text, "\n")
	if index == -1 {
		return text + "\n"
	}
	lastFragment := string(text[index:])
	if strings.TrimSpace(lastFragment) != "" {
		return text + "\n"
	}
	return text
}

func (s *multiCommandSession) drainStdout() {
	//read any outstanding output
	for {
		_, has, _ := s.readResponse(drainTimeoutMs, nil, "")
		if !has {
			return
		}
	}
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan bool)
	go func() {
		defer close(c)
		wg.Wait()
		c <- true
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func (s *multiCommandSession) shellInit() (err error) {
	if s.promptSequence != "" {
		if _, err = s.Run(s.promptSequence, nil, initTimeoutMs); err != nil {
			return err
		}
	}

	var ts = toolbox.AsString(time.Now().UnixNano())
	var waitGroup = &sync.WaitGroup{}
	waitGroup.Add(1)

	if s.config.Shell == defaultShell {
		s.promptSequence = "PS1=\"" + ts + "\\$\""
		s.shellPrompt = ts + "$"
		s.escapedShellPrompt = escapeInput(s.shellPrompt)
	}

	var listener Listener
	listener = func(stdout string, hasMore bool) {
		if !hasMore {
			waitGroup.Done()
		}
	}

	_, err = s.Run("", listener, initTimeoutMs, "$")

	waitTimeout(waitGroup, 60*time.Second)
	s.drainStdout()
	_, err = s.Run(s.promptSequence, nil, defaultTimeoutMs, "$")
	if s.closeIfError(err) {
		return err
	}
	for i := 0; i < 3; i++ {
		s.system, err = s.Run("uname -s", nil, initTimeoutMs)
		s.system = strings.ToLower(strings.TrimSpace(s.system))
		if strings.TrimSpace(s.system) != "" && !strings.Contains(s.system, "$") {
			break
		}
	}
	s.drainStdout()
	return nil
}

func (s *multiCommandSession) init() (err error) {
	s.session, err = s.service.client.NewSession()
	defer func() {
		if err != nil {
			s.service.client.Close()
		}
	}()
	s.stdOutput = make(chan string)
	s.stdError = make(chan string)
	for k, v := range s.config.EnvVariables {
		err = s.session.Setenv(k, v)
		if err != nil {
			return err
		}
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := s.session.RequestPty(s.config.Term, s.config.Rows, s.config.Columns, modes); err != nil {
		return err
	}

	if s.stdInput, err = s.session.StdinPipe(); err != nil {
		return err
	}

	stdout, err := s.start(s.config.Shell)
	if s.closeIfError(err) {
		return err
	}
	if err = checkNotFound(stdout); err != nil {
		return fmt.Errorf("failed to open %v shell, %v", s.config.Shell, err)
	}
	return s.shellInit()
}

func checkNotFound(output string) error {
	if strings.Contains(output, "not found") {
		return fmt.Errorf("failed run %s", output)
	}
	return nil
}

func newMultiCommandSession(service *service, config *SessionConfig, replayCommands *ReplayCommands, recordSession bool) (MultiCommandSession, error) {
	if config == nil {
		config = &SessionConfig{}
	}
	config.applyDefault()

	result := &multiCommandSession{
		service:        service,
		config:         config,
		running:        1,
		recordSession:  recordSession,
		replayCommands: replayCommands,
	}
	return result, result.init()
}

type notificationWindow struct {
	checkpoint  *time.Time
	listener    Listener
	elapsedMs   int
	stdout      string
	frequencyMs int
}

func (t *notificationWindow) flush() {
	if t.listener == nil {
		return
	}
	if t.stdout != "" {
		t.listener(t.stdout, true)
	}

	t.listener("", false)
}

func (t *notificationWindow) notify(stdout string) {
	var now = time.Now()
	if t.listener == nil {
		return
	}
	t.stdout += stdout
	t.elapsedMs += int(now.Sub(*t.checkpoint) / time.Millisecond)
	t.checkpoint = &now
	if t.elapsedMs > t.frequencyMs {
		t.listener(t.stdout, true)
		t.stdout = ""
		t.elapsedMs = 0
	}
}

func newNotificationWindow(listener Listener, frequencyMs int) *notificationWindow {
	var now = time.Now()
	return &notificationWindow{
		checkpoint:  &now,
		listener:    listener,
		frequencyMs: frequencyMs,
	}
}
