package multinet

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

func a() {
	fmt.Sprintf("")
}

type GroupTCPConn struct {
	sync.Mutex
	groupID int
	cap     int

	errorchannel   chan error
	writeChannel   chan *packageData
	tcpConn        map[int]*net.TCPConn
	virtualTCPConn map[int]*TCPConn

	listener *TCPListener
}

func newGroupTCPConn(groupID int, listener *TCPListener) *GroupTCPConn {
	groupTCPConn := new(GroupTCPConn)
	groupTCPConn.groupID = groupID

	groupTCPConn.errorchannel = make(chan error, 1024)
	groupTCPConn.writeChannel = make(chan *packageData, 1024)
	groupTCPConn.tcpConn = make(map[int]*net.TCPConn)
	groupTCPConn.virtualTCPConn = make(map[int]*TCPConn)

	groupTCPConn.listener = listener
	return groupTCPConn
}

func (self *GroupTCPConn) addConn(clientID int, conn *net.TCPConn) {
	//fmt.Printf("clientID %d\n", clientID)
	self.Lock()
	defer self.Unlock()
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
	self.Lock()
	defer self.Unlock()
	tcpConn = newTCPConn(self, 0)
	self.virtualTCPConn[tcpConn.syncID] = tcpConn
	return
}

func (self *GroupTCPConn) deleteTCPConn(syncID int) {
	self.Lock()
	defer self.Unlock()
	delete(self.virtualTCPConn, syncID)
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
