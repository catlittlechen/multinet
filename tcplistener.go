package multinet

import (
	"net"
	"strconv"
	"strings"
	"sync"
)

type clientIDGenerate struct {
	sync.Mutex
	id int
}

var cid = new(clientIDGenerate)

func getClientID() int {
	cid.Lock()
	defer cid.Unlock()
	cid.id++
	return cid.id
}

type TCPListener struct {
	sync.Mutex

	listener *net.TCPListener

	ifListen   bool
	tcpChannel chan *TCPConn
	errChannel chan error

	groupTCPConn map[int]GroupTCPConn
}

func ListenTCP(netStr string, laddr *net.TCPAddr) (*TCPListener, error) {
	tcpListener, err := net.ListenTCP(netStr, laddr)
	if err != nil {
		return nil, err
	}

	listener := new(TCPListener)
	listener.tcpChannel = make(chan *TCPConn, 1024)
	listener.errChannel = make(chan error, 1024)
	listener.listener = tcpListener
	listener.groupTCPConn = make(map[int]GroupTCPConn)

	return listener, nil
}

func (self *TCPListener) AcceptTCP() (tcpConn *TCPConn, err error) {

	go self.acceptTCP()
	select {
	case tcpConn = <-self.tcpChannel:
	case err = <-self.errChannel:
	}
	return
}

func (self *TCPListener) acceptTCP() {
	self.Lock()
	if self.ifListen {
		self.Unlock()
		return
	}
	self.ifListen = true
	defer func() {
		self.ifListen = false
	}()
	self.Unlock()

	data := make([]byte, 1024)
	groupID := 0
	for {
		conn, err := self.listener.AcceptTCP()
		if err != nil {
			self.errChannel <- err
			return
		}

		count, err := conn.Read(data)
		if err != nil {
			self.errChannel <- err
			return
		}

		dataStr := string(data[:count])
		clientID := getClientID()

		//new group
		if strings.Compare(dataStr, getGroupID) == 0 {
			groupID = getUniqueID()
			_, err = conn.Write([]byte(strconv.Itoa(groupID) + "&" + strconv.Itoa(clientID)))
			if err != nil {
				conn.Close()
				self.errChannel <- err
				return
			}

			groupTCPConn := newGroupTCPConn(groupID, self)
			groupTCPConn.addConn(clientID, conn)

			self.groupTCPConn[groupID] = *groupTCPConn
			continue
		}

		//new group member
		if groupID, err = strconv.Atoi(dataStr); err == nil {
			_, err = conn.Write([]byte(strconv.Itoa(groupID) + "&" + strconv.Itoa(clientID)))
			if err != nil {
				conn.Close()
				self.errChannel <- err
				return
			}
			groupTCPConn := self.groupTCPConn[groupID]
			groupTCPConn.addConn(clientID, conn)
		} else {
			conn.Close()
		}

	}
}

func (self *TCPListener) Close() error {
	return self.listener.Close()
}
