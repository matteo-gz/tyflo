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
	Connect(ctx context.Context, file, username, host string, port int) error
	ConnectByPassword(ctx context.Context, password, username, host string, port int) error
	ConnectByPrivateKey(ctx context.Context, privateKey, username, host string, port int) error
	Start(serverPort int) error
	Status() bool
	Close() error
}

func NewSshTunnel() SshI {
	return &SshImpl{}
}

type SshImpl struct {
	conn   *ssh.Client
	svc    *socks5.Server
	cancel context.CancelFunc
}

func (s *SshImpl) Status() bool {
	return true
}

func (s *SshImpl) Connect(ctx context.Context, file, username, host string, port int) (err error) {
	s.conn, err = ssh.NewClient(ctx, file, fmt.Sprintf("%v:%v", host, port), username)
	return err
}
func (s *SshImpl) ConnectByPrivateKey(ctx context.Context, privateKey, username, host string, port int) (err error) {
	s.conn, err = ssh.NewClientByPrivateKey(ctx, privateKey, fmt.Sprintf("%v:%v", host, port), username)
	return err
}
func (s *SshImpl) ConnectByPassword(ctx context.Context, password, username, host string, port int) (err error) {
	s.conn, err = ssh.NewClientByPassword(ctx, password, fmt.Sprintf("%v:%v", host, port), username)
	return err
}
func (s *SshImpl) Start(serverPort int) (err error) {
	l := logger.NewDefaultLogger()
	s.svc = socks5.NewServer(
		socks5.WithLogger(l),
		socks5.WithDialer(s.conn),
		socks5.WithAuthenticator(socks5.NoAuthenticator{}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	err = s.svc.Start(ctx, fmt.Sprintf(":%d", serverPort))
	return err
}

func (s *SshImpl) Close() error {
	var errs []error
	s.cancel()
	if s.svc != nil {
		errs = append(errs, s.svc.Stop())
	}
	if s.conn != nil {
		errs = append(errs, s.conn.Close())
	}
	return errors.Join(errs...)
}
