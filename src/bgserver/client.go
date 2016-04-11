package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"bgserver/network"
	"bgserver/common"
)

var (
	host		= flag.String("s", "127.0.0.1", "indicate the host's ip the client will connect")
	port		= flag.Int("p", "8080", "specify the service port of remote host")
)

func main() {
	flag.Parse()

	address := fmt.Sprintf("%s:%d", *host, *port)
	conn, err := network.Dial(address, network.WithTimeout(10 * time.Second))
	if err != nil {
		common.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()



}
