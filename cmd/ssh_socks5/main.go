package main

import (
	"context"
	"flag"
	"github.com/matteo-gz/tyflo/pkg/config"
	"github.com/matteo-gz/tyflo/pkg/io"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"github.com/matteo-gz/tyflo/pkg/protocol/ssh"
	"log"
)

type Conf struct {
	File string `yaml:"file"`
	Addr string `yaml:"addr"`
	User string `yaml:"user"`
}

func (c *Conf) get() interface{} {
	return *c
}

var flagConfig string

func main() {
	c := Conf{}
	flag.StringVar(&flagConfig, "conf", "", "config path, eg: -conf config.yaml")
	flag.Parse()
	log.Println("f", flagConfig)
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
	log.Println("c", c)
	sshc, err := ssh.NewClient(c.File, c.Addr, c.User)
	if err != nil {
		log.Println(err)
		return
	}
	l := logger.NewDefaultLogger()
	ss := socks5.NewServer(l, sshc)
	err = ss.Start(context.Background(), ":1079")
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("ok")
	io.NewBlocker().Block()
}
