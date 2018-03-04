package ssh

import (
	"bytes"
	"fmt"
	"github.com/lunixbochs/vtclean"
	"github.com/pkg/errors"
	"github.com/viant/toolbox"
	"golang.org/x/crypto/ssh"
	"io"
	"strings"
	"sync/atomic"
	"time"
)

const defaultShell = "/bin/bash"

const defaultTimeoutMs = 5000

//MultiCommandSession represents a multi command session
type MultiCommandSession interface {
	Run(command string, timeoutMs int, terminators ...string) (string, error)
	ShellPrompt() string
	System() string
	Close()
}

//multiCommandSession represents a multi command session
//a new command are send vi stdin
type multiCommandSession struct {
	replayCommands     *ReplayCommands
	recordSession      bool
	session            *ssh.Session
	stdOutput          chan string
	stdError           chan string
	stdInput           io.WriteCloser
	shellPrompt        string
	escapedShellPrompt string
	system             string
	running            int32
}

func (s *multiCommandSession) Run(command string, timeoutMs int, terminators ...string) (string, error) {
	s.drainStdout()
	var stdin = command + "\n"
	_, err := s.stdInput.Write([]byte(stdin))
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v, err: %v", command, err)
	}
	var output string
	output, _, err = s.readResponse(timeoutMs, terminators...)
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
	s.stdInput.Close()
	s.session.Close()

}

func (s *multiCommandSession) closeIfError(err error) bool {
	if err != nil {
		s.Close()
		return true
	}
	return false
}

func (s *multiCommandSession) init(shell string) (string, error) {
	reader, err := s.session.StdoutPipe()
	if err != nil {
		return "", err
	}
	go s.drain(reader, s.stdOutput)

	errReader, err := s.session.StderrPipe()
	if err != nil {
		return "", err
	}
	go s.drain(errReader, s.stdError)
	if shell == "" {
		shell = defaultShell
	}
	err = s.session.Start(shell)
	if err != nil {
		return "", err
	}
	var output string
	output, _, err = s.readResponse(defaultTimeoutMs)
	return output, err
}

func (s *multiCommandSession) drain(reader io.Reader, out chan string) {
	var written int64 = 0
	buf := make([]byte, 128*1024)
	for {
		writter := new(bytes.Buffer)
		if atomic.LoadInt32(&s.running) == 0 {
			return
		}

		bytesRead, readError := reader.Read(buf)
		if bytesRead > 0 {
			bytesWritten, writeError := writter.Write(buf[:bytesRead])
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
			out <- string(writter.Bytes())
		}
		if s.closeIfError(readError) {
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


func (s *multiCommandSession) hasTerminator(input string, terminators ...string) bool {
	escapedInput := escapeInput(input)

	var shellPrompt = s.shellPrompt
	if shellPrompt == "" {
		shellPrompt = "$"
	}
	if s.escapedShellPrompt == "" && s.shellPrompt != "" {
		s.escapedShellPrompt = escapeInput(s.shellPrompt)
	}



	if (s.escapedShellPrompt != "" && strings.HasSuffix(escapedInput, s.escapedShellPrompt) || strings.HasSuffix(input, s.shellPrompt)) {
		return true
	}

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

func (s *multiCommandSession) readResponse(timeoutMs int, terminators ...string) (out string, has bool, err error) {
	if timeoutMs == 0 {
		timeoutMs = defaultTimeoutMs
	}
	var done int32
	defer atomic.StoreInt32(&done, 1)
	var errOut string
	var hasOutput bool
outer:
	for {
		select {

		case o := <-s.stdOutput:
			out += o
			if s.hasTerminator(out, terminators...) && len(s.stdOutput) == 0 {
				break outer
			}
		case e := <-s.stdError:
			errOut += e
			if s.hasTerminator(errOut, terminators...) && len(s.stdOutput) == 0 {
				break outer
			}

		case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
			break outer
		}
	}
	if errOut != "" {
		err = errors.New(errOut)
	}

	if len(out) > 0 {
		hasOutput = true
		var lines = strings.Split(out, "\n")
		var escapedLines = make([]string, 0)
		for _, line := range lines {
			line = strings.Replace(line, "\r", "", 1)
			if line == s.shellPrompt {
				continue
			}
			escapedLines = append(escapedLines, line)
		}
		out = strings.Join(escapedLines, "\r\n")
	}
	return out, hasOutput, err
}

func (s *multiCommandSession) drainStdout() {
	//read any outstanding output
	for {
		_, has, _ := s.readResponse(10, "")
		if !has {
			return
		}
	}
}

func newMultiCommandSession(client *ssh.Client, config *SessionConfig, replayCommands *ReplayCommands, recordSession bool) (MultiCommandSession, error) {
	if config == nil {
		config = &SessionConfig{}
	}
	config.applyDefault()
	session, err := client.NewSession()
	defer func() {
		if err != nil {
			client.Close()
		}
	}()
	if err != nil {
		return nil, err
	}
	for k, v := range config.EnvVariables {
		err = session.Setenv(k, v)
		if err != nil {
			return nil, err
		}
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty(config.Term, config.Rows, config.Columns, modes); err != nil {
		return nil, err
	}
	var writer io.WriteCloser
	writer, err = session.StdinPipe()
	if err != nil {
		return nil, err
	}
	result := &multiCommandSession{
		session:        session,
		stdOutput:      make(chan string),
		stdError:       make(chan string),
		stdInput:       writer,
		running:        1,
		recordSession:  recordSession,
		replayCommands: replayCommands,
	}
	_, err = result.init(config.Shell)
	if result.closeIfError(err) {
		return nil, err
	}

	var ts = toolbox.AsString(time.Now().UnixNano())
	result.shellPrompt, err = result.Run("PS1=\"\\h:\\u"+ts+"\\$\"", 1000)
	if result.closeIfError(err) {
		return nil, err
	}
	result.escapedShellPrompt =  escapeInput(result.shellPrompt)
	result.system, err = result.Run("uname -s", 10000)
	result.system = strings.ToLower(result.system)
	return result, err
}
