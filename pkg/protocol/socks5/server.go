package socks5

import (
	"context"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"net"
	"sync"
	"time"
)

type Server struct {
	l    *net.TCPListener
	log  logger.Logger
	pool *sync.Pool
}

const (
	bufSize = 32 * 1024
	alive   = 180 * time.Second
)

func NewServer(l logger.Logger) *Server {
	pool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, bufSize)
		}}
	s := &Server{
		log:  l,
		pool: pool,
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
			sess := newSession(c, s.log, s.pool)
			go sess.handle(ctx)
		}
	}
}
