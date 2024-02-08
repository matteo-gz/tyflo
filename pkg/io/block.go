package io

type Blocker struct {
	c chan struct{}
}

func NewBlocker() *Blocker {
	return &Blocker{
		c: make(chan struct{}),
	}
}
func (b *Blocker) Block() {
	<-b.c
}
