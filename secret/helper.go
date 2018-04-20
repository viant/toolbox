package secret

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
	"time"
	//	"github.com/bgentry/speakeasy"
)

//ReadingCredentialTimeout represents max time for providing CredentialsFromLocation
var ReadingCredentialTimeout = time.Second * 45

var ReadUserAndPassword = func(timeout time.Duration) (user string, pass string, err error) {
	completed := make(chan bool)
	var reader = func() {
		defer func() {
			completed <- true
		}()

		var bytePassword, bytePassword2 []byte
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter Username: ")
		user, _ = reader.ReadString('\n')
		fmt.Print("Enter Password: ")
		bytePassword, err = terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			err = fmt.Errorf("failed to read password %v", err)
			return
		}
		fmt.Print("\nRetype Password: ")
		bytePassword2, err = terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			err = fmt.Errorf("failed to read password %v", err)
			return
		}
		password := string(bytePassword)
		if string(bytePassword2) != password {
			err = errors.New("password did not match")
		}
	}
	go reader()
	select {
	case <-completed:
	case <-time.After(timeout):
		err = fmt.Errorf("reading credential timeout")
	}
	user = strings.TrimSpace(user)
	pass = strings.TrimSpace(pass)
	return user, pass, err
}
