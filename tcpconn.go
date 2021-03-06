package multinet

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

var dialSignleMutex = new(sync.Mutex)

// TCPConn the virtual tcp conn in multinet
type TCPConn struct {
	groupConn    *groupTCPConn
	writeChannel chan *packageData
	readChannel  chan *packageData
	syncID       int
}

// DialTCP get TCPConn of multinet
func DialTCP(netStr string, laddr, raddr *net.TCPAddr) (*TCPConn, error) {
	dialSignleMutex.Lock()
	defer dialSignleMutex.Unlock()

	// only first dial will work
	gtc := getGroupConn(netStr, laddr, raddr)
	if gtc != nil {
		return gtc.getTCPConn(), nil
	}

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

	gid, cid, tmpData, err := splitData(string(data[:count]))
	if err != nil {
		conn.Close()
		return nil, err
	}

	gtc = newGroupTCPConn(gid, netStr, laddr, raddr, nil)
	gtc.addRealConn(cid, conn, tmpData)
	setGroupConn(netStr, laddr, raddr, gtc)

	go func() {
		for i := 1; i < initTCPCount; i++ {
			if err = gtc.dial(); err != nil {
				fmt.Println(err)
			}
		}
	}()

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

// Close the TCPConn
func (tc *TCPConn) Close() {
	tc.groupConn.deleteTCPConn(tc.syncID)
}

//TODO if possible, check the group to return error
// now block

// Read data from TCPConn
func (tc *TCPConn) Read() ([]byte, error) {
	pd := <-tc.readChannel
	if pd == nil {
		return nil, errors.New("read channel close")
	}
	return []byte(pd.Data), nil
}

// Write data into TCPConn
func (tc *TCPConn) Write(data []byte) {
	pd := newPackageData(tc.groupConn.groupID, tc.syncID, 0, data)
	tc.writeChannel <- pd
	go func() {
		if len(tc.writeChannel) > cap(tc.writeChannel)/20 && tc.groupConn.listener == nil {
			if err := tc.groupConn.dial(); err != nil {
				fmt.Println(err)
				return
			}
		}
	}()
}
