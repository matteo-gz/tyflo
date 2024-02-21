package main

import (
	"flag"
	"fmt"
	"github.com/matteo-gz/tyflo/pkg/config"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	// 构建代理地址
	// "http://username:password@proxyhost:proxyport"
	encodedPassword := url.QueryEscape(c.Password)
	fmt.Println(encodedPassword, c.Password)
	fmt.Println(encodedPassword == c.Password)
	proxyUrl, _ := url.Parse(
		fmt.Sprintf("http://%v:%v@%v", c.User, encodedPassword, c.Addr))

	// 创建Transport,并设置代理
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	// 创建Client,设置Transport
	client := &http.Client{
		Transport: transport,
	}

	// 构建请求
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	// 处理响应
	defer resp.Body.Close()
	dump, _ := httputil.DumpResponse(resp, true)
	fmt.Println(len(dump))

}
