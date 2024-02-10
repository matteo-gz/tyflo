package ssh

import (
	"context"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
)

type Client struct {
	c *ssh.Client
}

func NewClient(file, addr, user string) (c *Client, err error) {
	privateKey, err := os.ReadFile(file)
	if err != nil {
		return
	}
	key, err := ssh.ParseRawPrivateKey(privateKey)
	if err != nil {
		return
	}
	sig, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return
	}
	sshClient, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(sig),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}
	c = &Client{
		c: sshClient,
	}
	return
}
func (c *Client) DialContext(ctx context.Context, addr string) (conn net.Conn, err error) {
	log.Println("ssh addr", addr)
	conn, err = c.c.DialContext(ctx, "tcp", addr)
	return conn, err
}
