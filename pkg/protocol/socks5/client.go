package socks5

import (
	"context"
	"errors"
	"fmt"
	"net"
)

var ErrReplyFail = errors.New("reply fail")

func NewClient(serverAddress string) *Client {
	return &Client{serverAddress: serverAddress}
}

type Client struct {
	c             net.Conn
	serverAddress string
}

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
func (c *Client) negotiate() error {
	req := NewClientNegotiateReq()
	req.SetNoAuthenticationRequired()
	_, err := c.c.Write(req.Bytes())
	return err
}
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
