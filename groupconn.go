package multinet

import (
	"net"
	"sync"
)

type synID struct {
	sync.Mutex
	id int
}

var globalSynID = new(synID)

func getSynID() int {
	globalSynID.Lock()
	defer globalSynID.Unlock()
	globalSynID.id++
	return globalSynID.id
}

type GroupTCPConn struct {
	groupID        int
	cap            int
	tcpConn        map[int]*net.TCPConn
	virtualTCPConn map[int]*TCPConn
}

var globalMapGroupTCPConn = make(map[string]*TCPGroupConn)

func getGroupConn(netStr string, laddr, raddr *net.TCPAddr) *GroupTCPConn {
	key := netStr + "&" + laddr.String() + "&" + raddr.String()
	return globalMapGroupTCPConn[key]
}

func setGroupConn(netStr string, laddr, raddr *net.TCPAddr, tgc *GroupTCPConn) {
	key := netStr + "&" + laddr.String() + "&" + raddr.String()
	globalMapGroupTCPConn[key] = tgc
	return
}

func (self *GroupTCPConn) addConn(clientID int, conn *net.TCPConn) {
	self.cap++
	self.tcpConn[clientID] = conn
	return
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
	tcpConn = &TCPConn{
		groupConn: self,
		syncID:    getSynID(),
	}
	self.virtualTCPConn[tcpConn.syncID] = tcpConn
	return
}
