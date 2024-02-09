package main

import (
	"context"
	"github.com/matteo-gz/tyflo/api"
	"github.com/matteo-gz/tyflo/pkg/io"
	"time"
)

func main() {
	var c api.Client
	ctx := context.TODO()
	c.Dial(ctx)
	c.Forward(ctx)
	go func() {
		for {
			time.Sleep(time.Hour)
		}
	}()
	io.NewBlocker().Block()
}
