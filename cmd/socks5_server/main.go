package main

import (
	"context"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"log"
)

func main() {
	l := logger.NewDefaultLogger()
	ss := socks5.NewServer(l, socks5.DefaultDialer{})
	err := ss.Start(context.Background(), ":1079")
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("ok")
	ch := make(chan struct{})
	<-ch
}
