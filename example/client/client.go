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

	tcpConn, err := multinet.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("Fatal Error %s\n", err)
		return
	}

	data := []byte("Hello Multinet Server!")
	tcpConn.Write(data)
	data, err = tcpConn.Read()
	fmt.Println(string(data))
	fmt.Println(err)

	time.Sleep(1e10)
}
