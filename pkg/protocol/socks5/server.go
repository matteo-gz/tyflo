package socks5

import (
	"context"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"net"
)

type Server struct {
	l   *net.TCPListener
	log logger.Logger
}

func NewServer(l logger.Logger) *Server {
	return &Server{log: l}
}
func (s *Server) Start(ctx context.Context, addr string) (err error) {
	a, err := net.ResolveTCPAddr(tcp, addr)
	if err != nil {
		return err
	}
	l, err := net.ListenTCP(tcp, a)
	if err != nil {
		return err
	}
	s.l = l
	go s.accept(ctx)
	return nil
}
func (s *Server) accept(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c, err := s.l.Accept()
			if err != nil {
				continue
			}
			sess := newSession(c, s.log)
			go sess.handle(ctx)
		}
	}
}
