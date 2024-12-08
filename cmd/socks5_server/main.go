package main

import (
	"context"
	"flag"
	"log"

	"github.com/matteo-gz/tyflo/pkg/config"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
)

type UserAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Conf struct {
	Addr  string     `yaml:"addr"`
	Users []UserAuth `yaml:"users"`
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
	var auth socks5.Option = socks5.WithAuthenticator(socks5.NoAuthenticator{})
	if len(c.Users) > 0 {
		userMap := make(map[string]string)
		for _, u := range c.Users {
			userMap[u.Username] = u.Password
		}
		auth = socks5.WithAuthenticator(socks5.NewUserPassAuthenticator(userMap))
	}
	ss := socks5.NewServer(
		socks5.WithLogger(l),
		socks5.WithDialer(socks5.DefaultDialer{}),
		auth,
	)
	err = ss.Start(context.Background(), c.Addr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("ok")
	ch := make(chan struct{})
	<-ch
}
