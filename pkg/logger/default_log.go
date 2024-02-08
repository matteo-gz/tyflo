package logger

import (
	"context"
	"log"
)

type DefaultLog struct {
}

func NewDefaultLogger() *DefaultLog {
	return &DefaultLog{}
}
func (l *DefaultLog) DebugF(ctx context.Context, v ...interface{}) {

	log.Println("debug", v)
}

func (l *DefaultLog) ErrorF(ctx context.Context, v ...interface{}) {
	log.Println("error", v)
}
