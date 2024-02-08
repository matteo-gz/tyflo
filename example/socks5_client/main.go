package main

import (
	"context"
	"github.com/matteo-gz/tyflo/pkg/io"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"log"
)

func main() {
	go server()
	io.NewBlocker().Block()
}

func server() {
	l := logger.NewDefaultLogger()
	ss := socks5.NewServer(l)
	err := ss.Start(context.Background(), ":1079")
	if err != nil {
		log.Println(err)
	}
	log.Println("ok")
}
