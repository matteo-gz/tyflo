start:
	go run example/socks5_client/main.go
send:
	curl --proxy socks5://us:cn@127.0.0.1:1079 http://example.com/
