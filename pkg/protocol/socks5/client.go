package socks5

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/matteo-gz/tyflo/pkg/logger"
)

// ErrReplyFail 回复失败错误
var ErrReplyFail = errors.New("reply fail")

// NewClient 创建新的SOCKS5客户端
func NewClient(serverAddress string, l logger.Logger) *Client {
	return &Client{serverAddress: serverAddress, log: l}
}

// Client SOCKS5客户端结构体
type Client struct {
	c             net.Conn
	serverAddress string
	log           logger.Logger
}

// Dial 建立无认证的SOCKS5连接
func (c *Client) Dial(ctx context.Context, address string) (conn net.Conn, err error) {
	if c.c, err = dial(ctx, c.serverAddress); err != nil {
		return
	}
	if err = c.negotiate(); err != nil {
		return
	}
	if err = c.authenticate(ctx); err != nil {
		return
	}
	if err = c.handleRequest(ctx, address); err != nil {
		return
	}
	return c.c, nil
}

// DialWithUsernamePassword 使用用户名密码建立SOCKS5连接
func (c *Client) DialWithUsernamePassword(ctx context.Context, address, user, password string) (conn net.Conn, err error) {
	if c.c, err = dial(ctx, c.serverAddress); err != nil {
		c.log.DebugF(ctx, "dial")
		return
	}
	if err = c.negotiateWithUserPassword(); err != nil {
		c.log.DebugF(ctx, "negotiateWithUserPassword")
		return
	}
	if err = c.authenticateWithUserPassword(ctx, user, password); err != nil {
		c.log.DebugF(ctx, "authenticateWithUserPassword")
		return
	}
	c.log.DebugF(ctx, "handleRequest.before", address)
	if err = c.handleRequest(ctx, address); err != nil {
		c.log.DebugF(ctx, "handleRequest")
		return
	}
	return c.c, nil
}

// negotiate 进行无认证协商
func (c *Client) negotiate() error {
	req := NewClientNegotiateReq()
	req.SetNoAuthenticationRequired()
	_, err := c.c.Write(req.Bytes())
	return err
}

// negotiateWithUserPassword 进行用户名密码认证协商
func (c *Client) negotiateWithUserPassword() error {
	req := NewClientNegotiateReq()
	req.SetUsernamePassword()
	_, err := c.c.Write(req.Bytes())
	return err
}

// authenticate 进行无认证验证
func (c *Client) authenticate(ctx context.Context) error {
	r := NewServerNegotiateReply()
	if err := r.Decode(c.c); err != nil {
		return err
	}
	if r.Method == MethodNoAuthenticationRequired {
		return nil
	}
	return ErrMethodNotSupport
}

// authenticateWithUserPassword 进行用户名密码认证
func (c *Client) authenticateWithUserPassword(ctx context.Context, user, password string) error {
	r := NewServerNegotiateReply()
	if err := r.Decode(c.c); err != nil {
		c.log.DebugF(ctx, "decode")
		return err
	}
	switch r.Method {
	case MethodNoAuthenticationRequired:
		return nil
	case MethodUsernamePassword:
		req := NewUsernamePasswordReq()
		err := req.SetUsernamePassword(user, password)
		if err != nil {
			c.log.DebugF(ctx, "SetUsernamePassword")
			return err
		}
		c.log.DebugF(ctx, "NewUsernamePasswordReq")
		_, err = c.c.Write(req.Bytes())
		if err != nil {
			c.log.DebugF(ctx, "Write")
			return err
		}
		err = NewUsernamePasswordReply().Decode(c.c)
		c.log.DebugF(ctx, "NewUsernamePasswordReply.Decode", err)
		return err
	default:
		return fmt.Errorf("method%v %w", r.Method, ErrMethodNotSupport)
	}
}

// handleRequest 处理SOCKS5请求
func (c *Client) handleRequest(ctx context.Context, address string) (err error) {
	r := NewClientRequest()
	if err = r.SetCmdConnect(address); err != nil {
		return err
	}
	if _, err = c.c.Write(r.Bytes()); err != nil {
		return err
	}
	re := NewServerReply()
	if err = re.Decode(c.c); err != nil {
		return err
	}
	if re.REP != RepSucceeded {
		return fmt.Errorf("%w %v", ErrReplyFail, re.REP)
	}
	return nil
}
