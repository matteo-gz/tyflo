package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/matteo-gz/tyflo/pkg/config"
	io2 "github.com/matteo-gz/tyflo/pkg/io"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"io"
	"log"
	"net"
	"net/http"
)

type Conf struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Addr     string `yaml:"addr"`
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
	log.Println("c", c)
	l := logger.NewDefaultLogger()
	sc := socks5.NewClient(c.Addr, l)

	curl("https://www.example.com", sc, c.User, c.Password)
	io2.NewBlocker().Block()
}
func curl(url1 string, sc *socks5.Client, user, password string) {
	// 用网络连接创建Transport
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			fmt.Println("DialContext", network, addr)
			if network == "tcp" {
				conn, err := sc.DialWithUsernamePassword(context.Background(), addr, user, password)
				if err != nil {
					fmt.Println("DialContext.err", addr, err)
					return nil, err
				}
				return conn, nil
			}
			// return conn,nil
			return nil, errors.New("not tcp")
		},
	}

	// 用Transport创建Client
	client := &http.Client{Transport: transport}

	// 发起HTTP GET请求
	resp, err := client.Get(url1)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// 读取并打印响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("http", url1, len(body))
}
