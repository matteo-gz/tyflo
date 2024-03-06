package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/matteo-gz/tyflo/pkg/config"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"log"
	"net"
	"time"
)

type Conf struct {
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	Addr       string `yaml:"addr"`
	TargetAddr string `yaml:"target_addr"`
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
	var l logger.Logger
	l = logger.NewDefaultLogger()
	sc := socks5.NewClient(c.Addr, l)
	//cc, err := sc.Dial(context.Background(), c.TargetAddr)
	cc, err := sc.DialWithUsernamePassword(context.Background(), c.TargetAddr, c.User, c.Password)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	writeData(cc)
	for {
		select {
		case <-t.C:
			writeData(cc)
		}
	}
}
func writeData(cc net.Conn) {
	n, err := cc.Write([]byte{1})
	if err != nil {
		fmt.Println("write err", err)
	} else {
		fmt.Println("write n", n)
	}
}
