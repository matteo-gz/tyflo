package ssh

import "golang.org/x/crypto/ssh"

func NewServer() {
	ssh.NewServerConn(nil, nil)
}
