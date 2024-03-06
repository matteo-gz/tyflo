package main

import (
	"fmt"
	"net"
	"time"
)

func main() {

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

		// 处理连接
		go handleConn(conn)
	}
}
func handleConn(conn net.Conn) {
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println("read err", err)
			} else {
				fmt.Println("read n ", n)
			}
			fmt.Println("read data byte", buf)
			fmt.Println("read data string", string(buf))
			time.Sleep(2 * time.Second)
		}
	}()
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		writeData(conn) // when start
		for {
			select {
			case <-t.C:
				writeData(conn)
			}
		}
	}()
}
func writeData(conn net.Conn) {
	n, err := conn.Write([]byte{1})
	if err != nil {
		fmt.Println("write err", err)
	} else {
		fmt.Println("write n", n)
	}
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
