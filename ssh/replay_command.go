package ssh

import (
	"fmt"
	"github.com/viant/toolbox"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
)

//ReplayCommand represent a replay command
type ReplayCommand struct {
	Stdin  string
	Index  int
	Stdout []string
	Error  string
}

//replayCommands represnets command grouped by stdin
type ReplayCommands struct {
	Commands map[string]*ReplayCommand
	Keys     []string
	BaseDir  string
}

//Register register stdin and corresponding stdout conversation
func (c *ReplayCommands) Register(stdin, stdout string) {
	if _, has := c.Commands[stdin]; !has {
		c.Commands[stdin] = &ReplayCommand{
			Stdin:  stdin,
			Stdout: make([]string, 0),
		}
		c.Keys = append(c.Keys, stdin)
	}
	c.Commands[stdin].Stdout = append(c.Commands[stdin].Stdout, stdout)
}

//return stdout pointed by index and increases index or empty string if exhausted
func (c *ReplayCommands) Next(stdin string) string {
	var stdout = c.Commands[stdin].Stdout
	var index = c.Commands[stdin].Index
	if index < len(stdout) {
		c.Commands[stdin].Index++
		return stdout[index]
	}
	return ""
}

func (c *ReplayCommands) Enable(source interface{}) (err error) {
	switch value := source.(type) {
	case *service:
		value.replayCommands = c
		value.recordSession = true
	case *multiCommandSession:
		value.replayCommands = c
		value.recordSession = true
	default:
		err = fmt.Errorf("unsupported type: %T", source)
	}
	return err
}

func (c *ReplayCommands) Disable(source interface{}) (err error) {
	switch value := source.(type) {
	case *service:
		value.replayCommands = nil
		value.recordSession = false
	case *multiCommandSession:
		value.replayCommands = nil
		value.recordSession = false
	default:
		err = fmt.Errorf("unsupported type: %T", source)
	}
	return err
}

//Store stores replay command in the base directory
func (c *ReplayCommands) Store() error {
	err := toolbox.CreateDirIfNotExist(c.BaseDir)
	if err != nil {
		return err
	}
	for i, key := range c.Keys {
		var command = c.Commands[key]
		var filenamePrefix = path.Join(c.BaseDir, fmt.Sprintf("%03d", i+1))
		var stdinFilename = filenamePrefix + "_000.stdin"
		err := ioutil.WriteFile(stdinFilename, []byte(command.Stdin), 0644)
		if err != nil {
			return err
		}
		for j, stdout := range command.Stdout {
			var stdoutFilename = fmt.Sprintf("%v_%03d.stdout", filenamePrefix, j+1)
			err := ioutil.WriteFile(stdoutFilename, []byte(stdout), 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//Load loads replay command from base directory
func (c *ReplayCommands) Load() error {
	parent, err := os.Open(c.BaseDir)
	if err != nil {
		return err
	}
	files, err := parent.Readdir(1000)
	if err != nil {
		return err
	}
	var stdinMap = make(map[string]string)
	var stdoutMap = make(map[string]string)

	for _, candidate := range files {
		ext := path.Ext(candidate.Name())
		var contentMap map[string]string
		if ext == ".stdin" {
			contentMap = stdinMap
		} else if ext == ".stdout" {
			contentMap = stdoutMap
		} else {
			continue
		}
		var filename = path.Join(c.BaseDir, candidate.Name())
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil
		}
		contentMap[candidate.Name()] = string(content)
	}

	for key, stdin := range stdinMap {
		var prefix = key[:len(key)-10]
		var candidateKeys = toolbox.MapKeysToStringSlice(stdoutMap)
		sort.Strings(candidateKeys)
		for _, candidateKey := range candidateKeys {
			if strings.HasPrefix(candidateKey, prefix) {
				stdout := stdoutMap[candidateKey]
				c.Register(stdin, stdout)
			}
		}
	}
	return nil
}

//Shell returns command shell
func (c *ReplayCommands) Shell() string {
	for _, candidate := range c.Commands {
		if strings.HasPrefix(candidate.Stdin, "PS1=") && len(candidate.Stdout) > 0 {
			return candidate.Stdout[0]
		}
	}
	return ""
}

//System returns system name
func (c *ReplayCommands) System() string {
	for _, candidate := range c.Commands {
		if strings.HasPrefix(candidate.Stdin, "uname -s") && len(candidate.Stdout) > 0 {
			return strings.ToLower(candidate.Stdout[0])
		}
	}
	return ""
}

//NewReplayCommands create a new replay commands or error if provided basedir does not exists and can not be created
func NewReplayCommands(basedir string) (*ReplayCommands, error) {
	if !toolbox.FileExists(basedir) {
		err := os.MkdirAll(basedir, 0744)
		if err != nil {
			return nil, err
		}
	}
	return &ReplayCommands{
		Commands: make(map[string]*ReplayCommand),
		Keys:     make([]string, 0),
		BaseDir:  basedir,
	}, nil
}
