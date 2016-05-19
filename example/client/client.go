package main

import (
	"fmt"
	"github.com/catlittlechen/multinet"
	"net"
	"sync"
	"time"
)

var wg = new(sync.WaitGroup)

func main() {
	bindAddr := "127.0.0.1:10084"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", bindAddr)
	if err != nil {
		fmt.Printf("Fatal Error %s\n", err)
		return
	}
	startTime := time.Now().UnixNano()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go doTask(tcpAddr)
	}
	wg.Wait()
	endTime := time.Now().UnixNano()
	fmt.Println(endTime - startTime)
}

func doTask(tcpAddr *net.TCPAddr) {
	defer wg.Done()
	tcpConn, err := multinet.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("Fatal Error %s\n", err)
		return
	}
	defer tcpConn.Close()
	data := []byte("Hi Multinet Client!")
	for i := 0; i < 1000; i++ {
		tcpConn.Write(data)
		tcpConn.Read()
	}
	/*
		data, err = tcpConn.Read()
		fmt.Println(string(data))
		fmt.Println(err)
	*/
}
