package socks5

import (
	"context"
	"net"
)

func NewClient() *Client {
	return &Client{}
}

type Client struct {
	c net.Conn
}

func (c *Client) Dial(ctx context.Context, address string) error {
	conn, err := dial(ctx, address)
	if err != nil {
		return err
	}
	c.c = conn
	err = c.negotiate()
	if err != nil {
		return err
	}
	err = c.authenticate(ctx)
	if err != nil {
		return err
	}

	return nil
}
func (c *Client) negotiate() error {
	req := NewClientNegotiateReq()
	data := req.GetNoAuthenticationRequired()
	_, err := c.c.Write(data)
	return err
}
func (c *Client) authenticate(ctx context.Context) error {
	r := NewServerNegotiateReply()
	err := r.Decode(ctx, c.c)
	if err != nil {
		return err
	}
	if r.Method == MethodNoAuthenticationRequired {
		return nil
	}
	return ErrMethodNotSupport
}
func (c *Client) handleRequest(ctx context.Context) {
	// todo
}
func (c *Client) Forward(ctx context.Context) error {
	return nil
}
