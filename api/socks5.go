package api

import "context"

type Client interface {
	Dial(ctx context.Context) error
	Forward(ctx context.Context) error
}
type Server interface {
}