package socks5

import "context"

func New() *Client {
	return &Client{}
}

type Client struct {
}

func (c *Client) Dial(ctx context.Context) error {
	return nil
}
func (c *Client) Forward(ctx context.Context) error {
	return nil
}
