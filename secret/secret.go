package secret

import (
	"github.com/viant/toolbox/cred"
	"strings"
)

//Secrets represents CredentialsFromLocation secret map
type Secrets map[SecretKey]Secret

//NewSecrets creates new secrets
func NewSecrets(secrets map[string]string) Secrets {
	var result = make(map[SecretKey]Secret)
	if len(secrets) == 0 {
		return result
	}
	for k, v := range secrets {
		result[SecretKey(k)] = Secret(v)
	}
	return result
}

/**
SecretKey represent secret key
Take the following secrets as example:
<pre>

	"secrets": {
		"git": "${env.HOME}/.secret/git.json",
		"github.com": "${env.HOME}/.secret/github.json",
		"github.private.com": "${env.HOME}/.secret/github-private.json",
		"**replace**": "${env.HOME}/.secret/git.json",
	}

</pre>

The secret key can be static or dynamic. The first type is already enclosed with '*' or '#', the later is not.

In the command corresponding dynamic key can be enclosed with the following
'**' for password expansion  i.e.  command: **git** will expand to password from  git secret key
'##' for username expansion  i.e.  command: ##git## will expand to username from  git secret key
*/
type SecretKey string

//IsDynamic returns true if key is dynamic
func (s SecretKey) IsDynamic() bool {
	return !(strings.HasPrefix(string(s), "*") || strings.HasPrefix(string(s), "#"))
}

//String returns  secret key as string
func (s SecretKey) String() string {
	return string(s)
}

//Get extracts username or password or JSON based on key type (# prefix for user, otherwise password or JSON)
func (s SecretKey) Secret(cred *cred.Config) string {
	if strings.HasPrefix(s.String(), "#") || strings.HasSuffix(s.String(), ".username}") || strings.HasSuffix(s.String(), ".Username}") {
		return cred.Username
	}
	if cred.Password != "" {
		return cred.Password
	}
	return cred.Data
}

//Secret represents a secret
type Secret string

//IsLocation returns true if secret is a location
func (s Secret) IsLocation() bool {
	return !strings.ContainsAny(string(s), "{}[]=+()@#^&*|")
}
