package multinet

import (
	"errors"
	"fmt"
	"net"
)

type TCPConn struct {
	groupConn    *groupTCPConn
	writeChannel chan *packageData
	readChannel  chan *packageData
	syncID       int
}

func DialTCP(netStr string, laddr, raddr *net.TCPAddr) (*TCPConn, error) {

	// only first dial will work
	gtc := getGroupConn(netStr, laddr, raddr)
	if gtc != nil {
		return gtc.getTCPConn(), nil
	}
	fmt.Printf("Not Hit Cache")

	conn, err := net.DialTCP(netStr, laddr, raddr)
	if err != nil {
		return nil, err
	}
	conn.SetKeepAlive(true)

	_, err = conn.Write([]byte(getGroupID))
	if err != nil {
		conn.Close()
		return nil, err
	}

	data := make([]byte, 100)
	count, err := conn.Read(data)
	if err != nil {
		conn.Close()
		return nil, err
	}

	gid, cid, err := splitData(string(data[:count]))
	if err != nil {
		conn.Close()
		return nil, err
	}

	gtc = newGroupTCPConn(gid, netStr, laddr, raddr, nil)
	gtc.addConn(cid, conn)
	setGroupConn(netStr, laddr, raddr, gtc)

	for i := 1; i < initTCPCount; i++ {
		gtc.dial()
	}

	return gtc.getTCPConn(), nil
}

func newTCPConn(gtc *groupTCPConn, syncID int) *TCPConn {
	if syncID == 0 {
		syncID = getUniqueID()
	}
	return &TCPConn{
		groupConn:    gtc,
		writeChannel: gtc.writeChannel,
		readChannel:  make(chan *packageData, 1024),
		syncID:       syncID,
	}
}

func (tc *TCPConn) Close() {
	tc.groupConn.deleteTCPConn(tc.syncID)
}

func (tc *TCPConn) Read() ([]byte, error) {
	pd := <-tc.readChannel
	if pd == nil {
		return nil, errors.New("read channel close")
	}
	return []byte(pd.Data), nil
}

func (tc *TCPConn) Write(data []byte) {
	pd := newPackageData(tc.groupConn.groupID, tc.syncID, data)
	tc.writeChannel <- pd
}
