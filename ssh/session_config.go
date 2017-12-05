package ssh

//SessionConfig represents a new session config
type SessionConfig struct {
	EnvVariables map[string]string
	Shell        string
	Term         string
	Rows         int
	Columns      int
}

func (c *SessionConfig) applyDefault() {
	if c.Shell == "" {
		c.Shell = "/bin/bash"
	}
	if c.Term == "" {
		c.Term = "xterm"
	}
	if c.Rows == 0 {
		c.Rows = 100
	}
	if c.Columns == 0 {
		c.Columns = 100
	}
}
