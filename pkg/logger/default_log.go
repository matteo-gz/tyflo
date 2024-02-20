package logger

import (
	"context"
	"log"
	"strings"
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

type NopLog struct {
}

func NewNopLogLogger() *NopLog {
	return &NopLog{}
}
func (l *NopLog) DebugF(ctx context.Context, v ...interface{}) {

}

func (l *NopLog) ErrorF(ctx context.Context, v ...interface{}) {

}

type BufferLog struct {
	d []string
}

func NewBufferLogger() *BufferLog {
	return &BufferLog{}
}
func (b *BufferLog) Append(v string) {
	b.d = append(b.d, v)
}
func (b *BufferLog) String() string {
	return strings.Join(b.d, ",")
}
