package main

import (
	"context"
	"errors"
	"fmt"
	io2 "github.com/matteo-gz/tyflo/pkg/io"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	go server()
	curl("https://www.example.com")
	curl("https://www.baidu.com")
	io2.NewBlocker().Block()
}

func curl(url1 string) {
	// 用网络连接创建Transport
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			fmt.Println("DialContext", network, addr)
			if network == "tcp" {
				c := socks5.NewClient()
				err := c.Dial(ctx, ":1079", addr)
				if err != nil {
					return nil, err
				}
				return c.Forward(ctx), nil
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
		panic(err)
	}
	defer resp.Body.Close()

	// 读取并打印响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("http", url1, len(body))
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
