package ssh

import (
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

func New() (net.Conn, error) {
	sshClient, err := ssh.Dial("tcp", "localhost:22", &ssh.ClientConfig{
		User: "username",
		Auth: []ssh.AuthMethod{
			ssh.Password("password"),
		},
	})
	if err != nil {
		return nil, err
	}

	//todo
	//sshClient.ListenTCP()
	return &warp{c: sshClient}, nil
}

type warp struct {
	c *ssh.Client
}

func (w *warp) Close() error {
	return w.c.Close()
}

func (w *warp) LocalAddr() net.Addr {
	return w.c.LocalAddr()
}

func (w *warp) RemoteAddr() net.Addr {
	return w.c.RemoteAddr()
}

func (w *warp) Read(b []byte) (n int, err error) {
	panic("implement me")
}

func (w *warp) Write(b []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (w *warp) SetDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (w *warp) SetReadDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (w *warp) SetWriteDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}
