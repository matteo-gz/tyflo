package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/matteo-gz/tyflo/pkg/config"
	"github.com/matteo-gz/tyflo/pkg/logger"
	"github.com/matteo-gz/tyflo/pkg/protocol/socks5"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Conf struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Addr     string `yaml:"addr"`
	Count    int    `yaml:"count"`
	Batch    int    `yaml:"batch"`
	Debug    bool   `yaml:"debug"`
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
	var l logger.Logger

	if c.Debug {
		l = logger.NewDefaultLogger()
	} else {
		l = logger.NewNopLogLogger()
	}
	//sc := socks5.NewClient(c.Addr, l)
	csvHead()
	jobCount := c.Count
	jobBatch := c.Batch
	csvCount(jobBatch * jobCount)
	for i := 0; i < jobCount; i++ {
		job(l, c, jobBatch)
	}
	//"=SUBTOTAL(1,A1:A100)","=SUBTOTAL(1,B1:B100)/1000",,

	//io2.NewBlocker().Block()
}
func csvCount(total int) {
	bl := logger.NewBufferLogger()
	bl.Append(fmt.Sprintf(`"=SUBTOTAL(1,A3:A%d)"`, total+2))
	bl.Append(fmt.Sprintf(`"=SUBTOTAL(1,B3:B%d)/1000"`, total+2))
	bl.Append("")
	bl.Append("")
	fmt.Println(bl.String())
}
func csvHead() {
	bl := logger.NewBufferLogger()
	bl.Append(`dial time ms`)
	bl.Append(`curl time s`)
	bl.Append("url")
	bl.Append("body len")
	fmt.Println(bl.String())
}
func job(l logger.Logger, c Conf, total int) {
	wg := sync.WaitGroup{}
	for i := 0; i < total; i++ {
		sc := socks5.NewClient(c.Addr, l)
		wg.Add(1)
		go func() {
			defer wg.Done()
			curl("https://www.example.com", sc, c.User, c.Password)
		}()
	}
	wg.Wait()
}
func curl(url1 string, sc *socks5.Client, user, password string) {
	// 用网络连接创建Transport
	bl := logger.NewBufferLogger()
	defer func() {
		fmt.Println(bl.String())
	}()
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			//fmt.Println("DialContext", network, addr)
			if network == "tcp" {
				t1 := time.Now()
				conn, err := sc.DialWithUsernamePassword(context.Background(), addr, user, password)
				bl.Append(fmt.Sprintf("%d", time.Since(t1).Milliseconds()))
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
	t1 := time.Now()
	resp, err := client.Get(url1)
	bl.Append(fmt.Sprintf("%d", time.Since(t1).Milliseconds()))
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
	bl.Append(url1)
	bl.Append(fmt.Sprint(len(body)))
}
