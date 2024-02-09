package socks5

import (
	"context"
	"fmt"
	"net"
)

func NewClient(serverAddress string) *Client {
	return &Client{serverAddress: serverAddress}
}

type Client struct {
	c             net.Conn
	serverAddress string
}

func (c *Client) Dial(ctx context.Context, address string) (conn net.Conn, err error) {
	conn, err = dial(ctx, c.serverAddress)
	if err != nil {
		return
	}
	c.c = conn
	err = c.negotiate()
	if err != nil {
		return
	}
	err = c.authenticate(ctx)
	if err != nil {
		return
	}
	err = c.handleRequest(ctx, address)
	if err != nil {
		return
	}
	return
}
func (c *Client) negotiate() error {
	req := NewClientNegotiateReq()
	req.SetNoAuthenticationRequired()
	_, err := c.c.Write(req.Bytes())
	return err
}
func (c *Client) authenticate(ctx context.Context) error {
	r := NewServerNegotiateReply()
	err := r.Decode(c.c)
	if err != nil {
		return err
	}
	if r.Method == MethodNoAuthenticationRequired {
		return nil
	}
	return ErrMethodNotSupport
}
func (c *Client) handleRequest(ctx context.Context, address string) error {
	r := NewClientRequest()
	err := r.SetCmdConnect(address)
	if err != nil {
		return err
	}
	_, err = c.c.Write(r.Bytes())
	if err != nil {
		return err
	}
	re := NewServerReply()
	err = re.Decode(c.c)
	if err != nil {
		return err
	}
	if re.REP != RepSucceeded {
		return fmt.Errorf("reply:%v", re.REP)
	}
	return nil
}
