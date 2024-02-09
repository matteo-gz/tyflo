package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

func use() {
	// 用网络连接创建Transport
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// return conn,nil
			return nil, nil
		},
	}

	// 用Transport创建Client
	client := &http.Client{Transport: transport}

	// 发起HTTP GET请求
	resp, err := client.Get("https://www.example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 读取并打印响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}
