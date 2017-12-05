package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

//Tunnel represents a SSH forwarding link
type Tunnel struct {
	RemoteAddress string
	client        *ssh.Client
	Local         net.Listener
	Connections   []net.Conn
	mutex         *sync.Mutex
	closed        int32
}

func (f *Tunnel) tunnelTraffic(local, remote net.Conn) {
	defer local.Close()
	defer remote.Close()
	completionChannel := make(chan bool)
	go func() {
		_, err := io.Copy(local, remote)
		if err != nil {
			log.Printf("failed to copy remote to local: %v", err)
		}
		completionChannel <- true
	}()

	go func() {
		_, _ = io.Copy(remote, local)
		//if err != nil {
		//	log.Printf("failed to copy local to remote: %v", err)
		//}
		completionChannel <- true
	}()
	<-completionChannel
}

//Handle listen on local client to create tunnel with remote address.
func (f *Tunnel) Handle() error {
	for {
		if atomic.LoadInt32(&f.closed) == 1 {
			return nil
		}
		localclient, err := f.Local.Accept()
		if err != nil {
			return err
		}
		remote, err := f.client.Dial("tcp", f.RemoteAddress)
		if err != nil {
			return fmt.Errorf("failed to connect to remote: %v %v", f.RemoteAddress, err)
		}
		f.Connections = append(f.Connections, remote)
		f.Connections = append(f.Connections, localclient)
		go f.tunnelTraffic(localclient, remote)
	}
	return nil
}

//Close closes forwarding link
func (f *Tunnel) Close() error {
	atomic.StoreInt32(&f.closed, 1)
	_ = f.Local.Close()
	for _, remote := range f.Connections {
		_ = remote.Close()
	}
	return nil
}

//NewForwarding creates a new ssh forwarding link
func NewForwarding(client *ssh.Client, remoteAddress string, local net.Listener) *Tunnel {
	return &Tunnel{
		client:        client,
		RemoteAddress: remoteAddress,
		Connections:   make([]net.Conn, 0),
		Local:         local,
		mutex:         &sync.Mutex{},
	}
}
