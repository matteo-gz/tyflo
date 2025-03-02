package main

import (
	"context"
	"flag"
	"github.com/matteo-gz/tyflo/pkg/tunnel"
	"log"

	"github.com/matteo-gz/tyflo/pkg/config"
	"github.com/matteo-gz/tyflo/pkg/io"
)

type Conf struct {
	File    string `yaml:"file,omitempty"`
	SshHost string `yaml:"ssh_host"`
	SshPort int    `yaml:"ssh_port"`
	User    string `yaml:"user"`
	Port    int    `yaml:"port"`
	TypeX   string `yaml:"type"`
	Pass    string `yaml:"pass,omitempty"`
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st := tunnel.NewSshTunnel()
	if c.TypeX == "pass" {
		err = st.ConnectByPassword(ctx, c.Pass, c.User, c.SshHost, c.SshPort)
	} else if c.TypeX == "file" {
		err = st.Connect(ctx, c.File, c.User, c.SshHost, c.SshPort)
	} else {
		log.Println("type: file|pass")
		return
	}
	if err != nil {
		log.Println(err)
		return
	}
	defer st.Close()
	err = st.Start(c.Port)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("ok")
	io.NewBlocker().Block()
}
