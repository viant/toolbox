package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"path"
	"github.com/viant/toolbox"
	"github.com/viant/toolbox/cred"
)


func printUsage() {
	fmt.Printf("Usage: secret <secretname>\n where secretnmae if name of the file in $HOME/.secret/ directory\n i.e secret scp, will be generated $HOME/.secret/scp.json\n")
}

func main() {

	if len(os.Args) < 2 ||  os.Args[1] == "" {
		printUsage()
		return
	}

	var secretPath = path.Join(os.Getenv("HOME"), ".secret")
	if ! toolbox.FileExists(secretPath) {
		os.Mkdir(secretPath, 0744)
	}
	username, password := credentials()
	fmt.Println("")
	config := &cred.Config{
		Username:username,
		Password:password,
	}


	var privateKeyPAth = path.Join(os.Getenv("HOME"), ".ssh/id_rsa")
	if toolbox.FileExists(privateKeyPAth) {
		config.PrivateKeyPath = privateKeyPAth
	}
	var secretFile =  path.Join(secretPath, fmt.Sprintf("%v.json",os.Args[1]))
	err := config.Save(secretFile)
	if err != nil {
		log.Fatal(err)
	}
}

func credentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Failed to read password %v", err)
	}
	password := string(bytePassword)
	return strings.TrimSpace(username), strings.TrimSpace(password)
}