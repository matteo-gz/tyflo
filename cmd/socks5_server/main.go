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
	// 读取配置
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
	// 初始化日志
	l := logger.NewDefaultLogger()
	// 初始化认证
	var methods []socks5.Authenticator
	methods = append(methods, socks5.NoAuthenticator{})
	if len(c.Users) > 0 {
		log.Println("with auth")
		userMap := make(map[string]string)
		for _, u := range c.Users {
			userMap[u.Username] = u.Password
		}
		methods = append(methods, socks5.NewUserPassAuthenticator(userMap))

	} else {
		log.Println("without auth")
	}
	auth := socks5.WithAuthenticator(methods...)
	// 初始化socks5服务
	ss := socks5.NewServer(
		socks5.WithLogger(l),
		socks5.WithDialer(socks5.DefaultDialer{}),
		auth,
	)
	// 启动socks5服务
	err = ss.Start(context.Background(), c.Addr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("ok")
	// 等待退出
	ch := make(chan struct{})
	<-ch
}
