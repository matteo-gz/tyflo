package main

import (
	"context"
	"flag"
	"github.com/matteo-gz/tyflo/pkg/config"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"log"
)

type Conf struct {
	Addr string `yaml:"addr"`
}

var flagConfig string

func main() {
	c := Conf{}
	flag.StringVar(&flagConfig, "conf", "", "config path, eg: -conf config.yaml")
	flag.Parse()
	fn, err := config.Get(flagConfig)
	if err != nil {
		log.Println("get", err)
		return
	}
	err = fn(&c)
	if err != nil {
		log.Println("get2", err)
		return
	}
	log.Printf("c %#v", c)

	l := logger.NewDefaultLogger()
	ss := socks5.NewServer(l, socks5.DefaultDialer{})
	err = ss.Start(context.Background(), c.Addr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("ok")
	ch := make(chan struct{})
	<-ch
}
