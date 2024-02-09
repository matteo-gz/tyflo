package socks5

import (
	"context"
	"errors"
	"fmt"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"time"
)

var (
	ErrCacheType = errors.New("cache type err")
)

const dialTimeout = 5 * time.Second

type serverSession struct {
	c       *net.TCPConn
	log     logger.Logger
	address string
	buf     bufCache
}

type bufCache interface {
	Get() any
	Put(x any)
}

func newSession(c *net.TCPConn, l logger.Logger, b bufCache) *serverSession {
	return &serverSession{c: c, log: l, buf: b}
}
func (s *serverSession) config() {
	if err := s.c.SetKeepAlive(true); err != nil {
		_ = s.c.Close()
		return
	}
	if err := s.c.SetKeepAlivePeriod(alive); err != nil {
		_ = s.c.Close()
		return
	}
}
func (s *serverSession) handle(ctxP context.Context) {
	s.config()
	select {
	case <-ctxP.Done():
		return
	default:
		ctx := context.Background()
		//s.c.SetReadDeadline(time.Now()) todo set timeout
		clientVerReq, err := s.negotiate(ctx)
		if err != nil {
			s.log.ErrorF(ctx, "negotiate", clientVerReq, err)
			_ = s.c.Close()
			return
		}
		s.log.DebugF(ctx, fmt.Sprintf("clientVerReq:%#v", clientVerReq))
		if err = s.authenticate(ctx); err != nil {
			s.log.ErrorF(ctx, "authenticate", err)
			_ = s.c.Close()
			return
		}
		s.log.DebugF(ctx, "authenticate")
		clientRequest, err := s.handleRequest(ctx)
		if err != nil {
			s.log.ErrorF(ctx, "handleRequest", err)
			_ = s.c.Close()
			return
		}
		s.log.DebugF(ctx, "clientRequest", clientRequest)
		s.address = clientRequest.GetAddress()
		switch clientRequest.CMD {
		case CmdCONNECT:
			err = s.connect(ctx)
			if err != nil {
				_ = s.c.Close()
				s.log.ErrorF(ctx, "connect", err)
				return
			}
		}

	}
}

func (s *serverSession) negotiate(ctx context.Context) (req *ClientNegotiateReq, err error) {
	req = NewClientNegotiateReq()
	err = req.Decode(s.c)
	return
}

func (s *serverSession) authenticate(ctx context.Context) error {
	reply := NewServerNegotiateReply()
	reply.SetNotPassword()
	_, err := s.c.Write(reply.Bytes())
	return err
}
func (s *serverSession) handleRequest(ctx context.Context) (req *ClientRequest, err error) {
	req = NewClientRequest()
	err = req.Decode(s.c)
	return
}

func (s *serverSession) relay(ctx context.Context, dst, src io.ReadWriteCloser) {
	s.log.DebugF(ctx, "relay")
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		err := s.copy(ctx, dst, src)
		s.log.DebugF(ctx, "copy-done,src->dst", err)
		return err
	})
	eg.Go(func() error {
		err := s.copy(ctx, src, dst)
		s.log.DebugF(ctx, "copy-done,dst->src", err)
		return err
	})
	s.log.DebugF(ctx, "io wait")
	err := eg.Wait()
	if err != nil {
		s.log.ErrorF(ctx, "io done", err)
	} else {
		s.log.DebugF(ctx, "io done")
	}
	if err = dst.Close(); err != nil {
		s.log.ErrorF(ctx, "dst", err)
	}
	if err = src.Close(); err != nil {
		s.log.ErrorF(ctx, "src", err)
	}
}

func (s *serverSession) copy(ctx context.Context, dst io.Writer, src io.Reader) error {
	x := s.buf.Get()
	defer s.buf.Put(x)
	buf, ok := x.([]byte)
	if !ok {
		return ErrCacheType
	}
	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
func (s *serverSession) connect(ctx context.Context) error {
	reply := NewServerReply()
	reply.SetConnectDirectReply()
	n, err := s.c.Write(reply.Bytes())
	if err != nil {
		return err
	}
	s.log.DebugF(ctx, "NewServerReply", n)
	s.log.DebugF(ctx, "dial", s.address)
	conn, err := dial(ctx, s.address)
	if err != nil {
		return err
	}
	s.log.DebugF(ctx, "conn", conn.LocalAddr(), "\t", conn.RemoteAddr())
	s.log.DebugF(ctx, "source", s.c.LocalAddr(), "\t", s.c.RemoteAddr())

	go s.relay(ctx, conn, s.c)

	return nil

}
func dial(ctx context.Context, address string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout: dialTimeout,
	}
	return d.DialContext(ctx, tcp, address)
}
