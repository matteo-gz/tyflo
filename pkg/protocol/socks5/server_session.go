package socks5

import (
	"context"
	"fmt"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"net"
)

type serverSession struct {
	c   net.Conn
	log logger.Logger
}

func newSession(c net.Conn, l logger.Logger) *serverSession {
	return &serverSession{c: c, log: l}
}

func (s *serverSession) handle(ctxP context.Context) {
	defer func() {
		err := s.c.Close()
		if err != nil {
			s.log.ErrorF(ctxP, err)
		}
	}()
	select {
	case <-ctxP.Done():
		return
	default:
		ctx := context.Background()
		//s.c.SetReadDeadline(time.Now()) todo set timeout
		clientVerReq, err := s.negotiate(ctx)
		if err != nil {
			s.log.ErrorF(ctx, "negotiate", clientVerReq, err)
			return
		}
		s.log.DebugF(ctx, fmt.Sprintf("clientVerReq:%#v", clientVerReq))
		err = s.authenticate(ctx)
		if err != nil {
			s.log.ErrorF(ctx, "authenticate", err)
			return
		}
		s.log.DebugF(ctx, "authenticate")
		clientRequest, err := s.handleRequest(ctx)
		if err != nil {
			s.log.ErrorF(ctx, "handleRequest", err)
			return
		}
		s.log.DebugF(ctx, "clientRequest", clientRequest)
		switch clientRequest.CMD {
		case CmdCONNECT:
			s.connect(ctx)
		}

	}
}

func (s *serverSession) negotiate(ctx context.Context) (req *ClientNegotiateReq, err error) {
	req = NewClientNegotiateReq()
	err = req.Get(s.c)
	return
}

func (s *serverSession) authenticate(ctx context.Context) error {
	reply := NewServerNegotiateReply()
	reply.SetNotPassword()
	_, err := s.c.Write(reply.Get())
	return err
}
func (s *serverSession) handleRequest(ctx context.Context) (req *ClientRequest, err error) {
	req = NewClientRequest()
	err = req.Get(ctx, s.c)
	return
}
func (s *serverSession) relay(ctx context.Context) {
	r := NewServerReply()
	r.Set()
}
func (s *serverSession) connect(ctx context.Context) {
	go s.relay(ctx)
}
