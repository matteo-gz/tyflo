package ssh

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	c    *ssh.Client
	addr string
	conf *ssh.ClientConfig
}

func NewClientByPrivateKey(ctx context.Context, file, addr, user string) (c *Client, err error) {
	key, err := ssh.ParseRawPrivateKey([]byte(file))
	if err != nil {
		return
	}
	sig, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return
	}
	c = &Client{
		addr: addr,
		conf: &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(sig),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         contextToTimeout(ctx),
		},
	}
	err = c.connect()
	if err != nil {
		return nil, err
	}
	go c.retry(ctx)
	return
}
func NewClient(ctx context.Context, file, addr, user string) (c *Client, err error) {
	privateKey, err := os.ReadFile(file)
	if err != nil {
		return
	}
	return NewClientByPrivateKey(ctx, string(privateKey), addr, user)
}
func contextToTimeout(ctx context.Context) time.Duration {
	timeout := 30 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	return timeout
}
func NewClientByPassword(ctx context.Context, pass, addr, user string) (c *Client, err error) {

	c = &Client{
		addr: addr,
		conf: &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.Password(pass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         contextToTimeout(ctx),
		},
	}
	err = c.connect()
	if err != nil {
		return nil, err
	}
	go c.retry(ctx)
	return
}
func (c *Client) connect() error {
	sshClient, err := ssh.Dial("tcp", c.addr, c.conf)
	if err != nil {
		return err
	}
	c.c = sshClient
	return nil
}
func (c *Client) retry(ctx context.Context) {
	ch := make(chan struct{}, 1)
	for {
		select {
		case <-ctx.Done():
			log.Println("ssh client disconnected,give up retry.")
			return
		case <-ch:
			if err := c.connect(); err != nil {
				log.Println("connect retry", err)
				time.Sleep(10 * time.Second)
				ch <- struct{}{}
			} else {
				log.Println("ssh recovery")
			}
		default:
			log.Println("ssh holding")
			err := c.c.Wait()
			if err != nil {
				log.Println("ssh client err", err)
			} else {
				log.Println("ssh client leave")
			}
			ch <- struct{}{}
		}
	}
}
func (c *Client) DialContext(ctx context.Context, addr string) (conn net.Conn, err error) {
	log.Println("ssh addr", addr)
	conn, err = c.c.DialContext(ctx, "tcp", addr)
	return conn, err
}
func (c *Client) Close() error {
	return c.c.Close()
}
