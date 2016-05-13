package multinet

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

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
	self.Lock()
	defer self.Unlock()
	self.cap++
	self.tcpConn[clientID] = conn
	go self.read(clientID, conn)
	go self.write(clientID, conn)
	return
}

func (self *GroupTCPConn) read(clientID int, conn *net.TCPConn) {
	data := make([]byte, 1024)
	tmpData := make([]byte, 0)
	nowDataLen := 0
	var n int
	var err error
	for {
		n, err = conn.Read(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		tmpData = append(tmpData, data[:n]...)

	DealWithTmpData:
		if nowDataLen == 0 {
			if len(tmpData) < 4 {
				continue
			} else {
				nowDataLen, err = strconv.Atoi(string(tmpData[:4]))
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
		if nowDataLen > len(tmpData) {
			continue
		}

		pd := getPackageData()
		err = pd.Decode(tmpData[4:nowDataLen])
		if err != nil {
			fmt.Println(err)
			return
		}

		tmpData = tmpData[nowDataLen:]
		nowDataLen = 0

		if pd.GroupID != self.groupID {
			continue
		}
		if tcpConn, ok := self.virtualTCPConn[pd.SyncID]; ok {
			tcpConn.readChannel <- pd
		} else if self.listener != nil {
			tcpConn = newTCPConn(self, pd.SyncID)
			self.listener.tcpChannel <- tcpConn
			tcpConn.readChannel <- pd
		} else {
			putPackageData(pd)
		}

		if len(tmpData) != 0 {
			goto DealWithTmpData
		}
	}
}

func (self *GroupTCPConn) write(clientID int, conn *net.TCPConn) {
	var err error
	for {
		pd := <-self.writeChannel
		data := pd.Encode()
		length := strconv.Itoa(len(data) + 4)
		for len(length) < 4 {
			length = "0" + length
		}
		data = append([]byte(length), data...)
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		//TODO control stream
		putPackageData(pd)
	}
}

func (self *GroupTCPConn) Close() error {
	for _, tcpconn := range self.tcpConn {
		if err := tcpconn.Close(); err != nil {
			fmt.Println(err)
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
