package socks5

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/matteo-gz/tyflo/pkg/logger"
)

type Server struct {
	l             *net.TCPListener
	log           logger.Logger
	pool          *sync.Pool
	dialer        Dialer
	authenticator Authenticator
}

const (
	bufSize = 32 * 1024
	alive   = 180 * time.Second
)

type Option func(*Server)

func WithLogger(l logger.Logger) Option {
	return func(s *Server) {
		s.log = l
	}
}

func WithDialer(d Dialer) Option {
	return func(s *Server) {
		s.dialer = d
	}
}

func WithAuthenticator(a Authenticator) Option {
	return func(s *Server) {
		s.authenticator = a
	}
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, bufSize)
			},
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Server) Start(ctx context.Context, addr string) (err error) {
	a, err := net.ResolveTCPAddr(tcp, addr)
	if err != nil {
		return err
	}
	if s.l, err = net.ListenTCP(tcp, a); err != nil {
		return err
	}
	go s.accept(ctx)
	return nil
}

func (s *Server) accept(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c, err := s.l.AcceptTCP()
			if err != nil {
				s.log.ErrorF(ctx, "AcceptTCP", err)
				continue
			}
			s.log.DebugF(ctx, "newSession")
			sess := newSession(c, s.log, s.pool, s.dialer)
			go sess.handle(ctx)
		}
	}
}
