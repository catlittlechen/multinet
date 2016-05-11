package multinet

import (
	"encoding/json"
	"net"
)

type GroupTCPConn struct {
	groupID int
	cap     int

	errorchannel   chan error
	writeChannel   chan *packageData
	tcpConn        map[int]*net.TCPConn
	virtualTCPConn map[int]*TCPConn

	listener *TCPListener
}

func newGroupTCPConn(groupID int, listener *TCPListener) *GroupTCPConn {
	return &GroupTCPConn{
		groupID: groupID,

		errorchannel:   make(chan error, 1024),
		writeChannel:   make(chan *packageData, 1024),
		tcpConn:        make(map[int]*net.TCPConn),
		virtualTCPConn: make(map[int]*TCPConn),

		listener: listener,
	}
}

func (self *GroupTCPConn) addConn(clientID int, conn *net.TCPConn) {
	self.cap++
	self.tcpConn[clientID] = conn
	go self.read(clientID, conn)
	go self.write(clientID, conn)
	return
}

func (self *GroupTCPConn) read(clientID int, conn *net.TCPConn) {
	decoder := json.NewDecoder(conn)
	var err error
	for {
		pd := getPackageData()
		err = decoder.Decode(pd)
		if err != nil {
			return
		}
		if pd.GroupID != self.groupID {
			continue
		}
		if tcpConn, ok := self.virtualTCPConn[pd.SynID]; ok {
			tcpConn.readChannel <- pd
		} else if self.listener != nil {
			tcpConn = newTCPConn(self, pd.SynID)
			self.listener.tcpChannel <- tcpConn
			tcpConn.readChannel <- pd
		}
	}
}

func (self *GroupTCPConn) write(clientID int, conn *net.TCPConn) {
	encoder := json.NewEncoder(conn)
	var err error
	for {
		pd := <-self.writeChannel
		err = encoder.Encode(pd)
		if err != nil {
			return
		}
		//TODO control stream
		putPackageData(pd)
	}
}

func (self *GroupTCPConn) Close() error {
	for _, tcpconn := range self.tcpConn {
		if err := tcpconn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (self *GroupTCPConn) getTCPConn() (tcpConn *TCPConn) {
	tcpConn = newTCPConn(self, 0)
	self.virtualTCPConn[tcpConn.syncID] = tcpConn
	return
}

var globalMapGroupTCPConn = make(map[string]*GroupTCPConn)

func getGroupConn(netStr string, laddr, raddr *net.TCPAddr) *GroupTCPConn {
	key := netStr + "&" + laddr.String() + "&" + raddr.String()
	return globalMapGroupTCPConn[key]
}

func setGroupConn(netStr string, laddr, raddr *net.TCPAddr, tgc *GroupTCPConn) {
	key := netStr + "&" + laddr.String() + "&" + raddr.String()
	globalMapGroupTCPConn[key] = tgc
	return
}
