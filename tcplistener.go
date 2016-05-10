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
	listener     *net.TCPListener
	groupTCPConn map[int]GroupTCPConn
	groupID      int
}

func ListenTCP(netStr string, laddr *net.TCPAddr) (*TCPListener, error) {
	tcpListener, err := net.ListenTCP(netStr, laddr)
	if err != nil {
		return nil, err
	}

	listener := new(TCPListener)
	listener.listener = tcpListener
	listener.groupTCPConn = make(map[int]GroupTCPConn)

	return listener, nil
}

func (self *TCPListener) getGroupID() int {
	self.Lock()
	defer self.Unlock()

	self.groupID++
	return self.groupID
}

func (self *TCPListener) AcceptTCP() (*TCPConn, error) {

	data := make([]byte, 1024)
	groupID := 0
	for {

		conn, err := self.listener.AcceptTCP()
		if err != nil {
			return nil, err
		}

		count, err := conn.Read(data)
		if err != nil {
			return nil, err
		}

		dataStr := string(data[:count])
		clientID := getClientID()

		//new group
		if strings.Compare(dataStr, getGroupID) == 0 {
			groupID = self.getGroupID()
			_, err = conn.Write([]byte(strconv.Itoa(groupID) + "&" + strconv.Itoa(clientID)))
			if err != nil {
				conn.Close()
				return nil, err
			}
			groupTCPConn := &GroupTCPConn{
				groupID:        groupID,
				tcpConn:        make(map[int]*net.TCPConn),
				virtualTCPConn: make(map[int]*TCPConn),
			}
			groupTCPConn.addConn(clientID, conn)
			self.groupTCPConn[groupID] = *groupTCPConn
			return groupTCPConn.getTCPConn(), nil
		}

		//new group member
		if groupID, err = strconv.Atoi(dataStr); err == nil {
			_, err = conn.Write([]byte(strconv.Itoa(groupID) + "&" + strconv.Itoa(clientID)))
			if err != nil {
				conn.Close()
				return nil, err
			}
			groupTCPConn := self.groupTCPConn[groupID]
			groupTCPConn.addConn(clientID, conn)
		} else {
			conn.Close()
		}

	}
	return nil, nil
}

func (self *TCPListener) Close() error {
	return self.listener.Close()
}
