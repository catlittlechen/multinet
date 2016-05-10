package main

import (
	"fmt"
	"github.com/catlittlechen/multinet"
	"net"
	"time"
)

func main() {
	bindAddr := "127.0.0.1:10084"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", bindAddr)
	if err != nil {
		fmt.Printf("Fatal Error %s\n", err)
		return
	}

	_, err = multinet.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("Fatal Error %s\n", err)
		return
	}

	time.Sleep(1e10)
}
