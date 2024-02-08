package logger

import "context"

type Logger interface {
	DebugF(ctx context.Context, v ...interface{})
	ErrorF(ctx context.Context, v ...interface{})
}
