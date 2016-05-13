package main

import (
	"fmt"
	"github.com/catlittlechen/multinet"
	"net"
)

func main() {
	bindAddr := "127.0.0.1:10084"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", bindAddr)
	if err != nil {
		fmt.Printf("ResolveTCPAddr[%s] error[%s]\n", bindAddr, err)
		return
	}
	conn, err := multinet.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Printf("listenTCP[%s] error[%s]\n", tcpAddr, err)
		return
	}
	for {
		tcpConn, err := conn.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go dealTCPConn(tcpConn)

	}
}

func dealTCPConn(tcpConn *multinet.TCPConn) {
	data, _ := tcpConn.Read()
	data = []byte("Hi Multinet Client!")
	tcpConn.Write(data)
}
