package tunnel

import (
	"context"
	"errors"
	"fmt"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"github.com/matteo-gz/tyflo/pkg/protocol/ssh"
)

type SshI interface {
	Connect(file, username, host string, port int) error
	ConnectByPassword(password, username, host string, port int) error
	Start(serverPort int) error
	Close() error
}

func NewSshTunnel() SshI {
	return &SshImpl{}
}

type SshImpl struct {
	conn *ssh.Client
	svc  *socks5.Server
}

func (s *SshImpl) Connect(file, username, host string, port int) (err error) {
	s.conn, err = ssh.NewClient(file, fmt.Sprintf("%v:%v", host, port), username)
	return err
}
func (s *SshImpl) ConnectByPassword(password, username, host string, port int) (err error) {
	s.conn, err = ssh.NewClientByPassword(password, fmt.Sprintf("%v:%v", host, port), username)
	return err
}
func (s *SshImpl) Start(serverPort int) (err error) {
	l := logger.NewDefaultLogger()
	s.svc = socks5.NewServer(
		socks5.WithLogger(l),
		socks5.WithDialer(s.conn),
		socks5.WithAuthenticator(socks5.NoAuthenticator{}),
	)
	err = s.svc.Start(context.Background(), fmt.Sprintf(":%d", serverPort))
	return err
}

func (s *SshImpl) Close() error {
	var errs []error
	if s.conn != nil {
		errs = append(errs, s.conn.Close())
	}
	if s.svc != nil {
		errs = append(errs, s.svc.Stop())
	}
	_ = s.conn.Close()
	return errors.Join(errs...)
}
