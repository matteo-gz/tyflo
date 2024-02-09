package socks5

import (
	"context"
	"fmt"
	"net"
)

func NewClient() *Client {
	return &Client{}
}

type Client struct {
	c      net.Conn
	target string
}

func (c *Client) Dial(ctx context.Context, serverAddress, address string) error {
	conn, err := dial(ctx, serverAddress)
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
	err = c.handleRequest(ctx, address)
	if err != nil {
		return err
	}
	c.target = address
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
func (c *Client) handleRequest(ctx context.Context, address string) error {
	r := NewClientRequest()
	err := r.Decode(address)
	if err != nil {
		return err
	}
	_, err = c.c.Write(r.bytes())
	if err != nil {
		return err
	}
	re := NewServerReply()
	err = re.decode(c.c)
	if err != nil {
		return err
	}
	if re.REP != RepSucceeded {
		return fmt.Errorf("reply:%v", re.REP)
	}
	return nil
}
func (c *Client) Forward(ctx context.Context) net.Conn {
	return c.c
}
