package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/matteo-gz/tyflo/pkg/config"
	"github.com/matteo-gz/tyflo/pkg/io"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"github.com/matteo-gz/tyflo/pkg/protocol/ssh"
	"log"
)

type Conf struct {
	File  string `yaml:"file,omitempty"`
	Addr  string `yaml:"addr"`
	User  string `yaml:"user"`
	Port  int    `yaml:"port"`
	TypeX string `yaml:"type"`
	Pass  string `yaml:"pass,omitempty"`
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
	var (
		sshc *ssh.Client
	)
	if c.TypeX == "pass" {
		sshc, err = ssh.NewClientByPassword(c.Pass, c.Addr, c.User)
	} else if c.TypeX == "file" {
		sshc, err = ssh.NewClient(c.File, c.Addr, c.User)
	} else {
		log.Println("type: file|pass")
		return
	}
	if err != nil {
		log.Println(err)
		return
	}
	l := logger.NewDefaultLogger()
	ss := socks5.NewServer(l, sshc)
	err = ss.Start(context.Background(), fmt.Sprintf(":%d", c.Port))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("ok")
	io.NewBlocker().Block()
}
