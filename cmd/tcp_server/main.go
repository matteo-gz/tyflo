package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/matteo-gz/tyflo/pkg/config"
)

type Conf struct {
	Addr      string `yaml:"addr"`
	ReplyTime uint   `yaml:"reply_time"`
	Type      string `yaml:"type"`
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
	// 监听localhost:8080
	ln, err := net.Listen("tcp", ":3306")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer ln.Close()

	fmt.Println("Listening on localhost:3306")

	// 接受客户端连接
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}
		switch c.Type {
		case "tcp":
			go handleConn(conn, c.ReplyTime)
		case "http":
			go handleConnV1(conn)
		}
	}
}

func handleConn(conn net.Conn, ReplyTime uint) {
	ch := make(chan []byte, 40)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				log.Println("read done")
				return
			default:
				n, err := conn.Read(buf)
				if err != nil {
					log.Println("read err:", err)
					cancel()
					return
				} else {
					ch <- buf[:n]
				}
				log.Printf("read data n:%v err:%v byte %v \n", n, err, buf[:n])
				time.Sleep(2 * time.Second)
			}
		}
	}()
	go func() {
		t := time.NewTicker(time.Duration(ReplyTime) * time.Second)
		defer t.Stop()
		err := writeData(conn, []byte{0}) // when start
		if err != nil {
			cancel()
			return
		}
		for {
			select {
			case msg := <-ch:
				buf := make([]byte, len(msg))
				for i, v := range msg {
					buf[i] = math.MaxUint8 - v
				}
				err = writeData(conn, buf)
				if err != nil {
					log.Println("write err:", err)
					cancel()
					return
				}
			case <-t.C:
				err = writeData(conn, []byte{0})
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()
}
func writeData(conn net.Conn, data []byte) error {
	n, err := conn.Write(data)
	if err != nil {
		log.Println("write err", err)
	} else {
		log.Println("write n", n, data)
	}
	return err
}

// 处理单个连接
func handleConnV1(conn net.Conn) {
	defer conn.Close()

	fmt.Println("New connection established")

	// 读取请求
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		return
	}

	// 打印请求内容
	req := string(buf[:n])
	fmt.Println("Request: ", req)

	// 写回响应
	_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nHello World!\n"))
	if err != nil {
		fmt.Println("Error writing: ", err.Error())
	}
}
