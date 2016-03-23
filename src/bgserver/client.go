package main

import (
	"flag"
	"fmt"
	"net"
)

var (
	host		= flag.String("s", "127.0.0.1", "indicate the host's ip the client will connect")
	port		= flag.Int("p", "8080", "specify the service port of remote host")
)

func main() {
	flag.Parse()


}
