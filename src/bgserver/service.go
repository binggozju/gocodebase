package main

import (
	"flag"
	"fmt"

	"tcpserver/network"
)

var (
	listen_ip		= flag.String("s", "127.0.0.1", "the ip the server will listen")
	listen_port		= flag.Int("p", 8080, "the port the server will listen")
)

func main() {
	flag.Parse()

	s := network.TCPServer(*listen_ip, *listen_port)
	fmt.Printf("The TCP server is listenning at %s:%d\n", *listen_ip, *port)
	s.Run()
}
