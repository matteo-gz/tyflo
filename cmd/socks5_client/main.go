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
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Addr        string `yaml:"addr"`
	TargetAddr  string `yaml:"target_addr"`
	RequestTime uint   `yaml:"request_time"`
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
		log.Println("err", err)
		return
	}
	t := time.NewTicker(time.Duration(c.RequestTime) * time.Second)
	defer t.Stop()
	writeData(cc)
	go readData(cc)
	for {
		select {
		case <-t.C:
			writeData(cc)
		}
	}
}
func readData(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("read  err:%v \n", err)
		} else {
			log.Printf("read n:%v %v \n", n, buf[:n])
		}
		time.Sleep(1 * time.Second)
	}
}
func writeData(cc net.Conn) {
	var data []byte
	for _, v := range fmt.Sprint(time.Now().Unix()) {
		b := byte(v - '0')
		data = append(data, b)
	}
	n, err := cc.Write(data)
	if err != nil {
		log.Println("write err", err)
	} else {
		log.Println("write n", n, data)
	}
}
