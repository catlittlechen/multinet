package multinet

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

type TCPConn struct {
	groupConn    *GroupTCPConn
	writeChannel chan *packageData
	readChannel  chan *packageData
	syncID       int
}

func DialTCP(netStr string, laddr, raddr *net.TCPAddr) (*TCPConn, error) {

	// only first dial will work
	groupTCPConn := getGroupConn(netStr, laddr, raddr)
	if groupTCPConn != nil {
		return groupTCPConn.getTCPConn(), nil
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

	data := make([]byte, 1024)
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

	groupTCPConn = newGroupTCPConn(gid, nil)
	groupTCPConn.addConn(cid, conn)

	go func() {
		for i := 1; i < TCPCount; i++ {

			conn, err = net.DialTCP(netStr, laddr, raddr)
			if err != nil {
				return
			}

			_, err = conn.Write([]byte(strconv.Itoa(groupTCPConn.groupID)))
			if err != nil {
				conn.Close()
				return
			}
			count, err = conn.Read(data)
			if err != nil {
				conn.Close()
				return
			}
			gid, cid, err = splitData(string(data[:count]))
			if err != nil {
				conn.Close()
				return
			}

			if groupTCPConn.groupID == gid {
				groupTCPConn.addConn(cid, conn)
			} else {
				conn.Close()
				return
			}
		}
	}()

	setGroupConn(netStr, laddr, raddr, groupTCPConn)
	return groupTCPConn.getTCPConn(), nil
}

func newTCPConn(groupTCPConn *GroupTCPConn, syncID int) *TCPConn {
	if syncID == 0 {
		syncID = getUniqueID()
	}
	return &TCPConn{
		groupConn:    groupTCPConn,
		writeChannel: groupTCPConn.writeChannel,
		readChannel:  make(chan *packageData, 1024),
		syncID:       syncID,
	}
}

func (self *TCPConn) Close() {
	self.groupConn.deleteTCPConn(self.syncID)
}

func (self *TCPConn) Read() ([]byte, error) {
	pd := <-self.readChannel
	if pd == nil {
		return nil, errors.New("read channel close")
	}
	return []byte(pd.Data), nil
}

func (self *TCPConn) Write(data []byte) {
	pd := getPackageData()
	pd.GroupID = self.groupConn.groupID
	pd.SynID = self.syncID
	pd.Data = string(data)
	self.writeChannel <- pd
	//TODO control stream
}
