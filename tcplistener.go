package multinet

import (
	"net"
	"strconv"
	"strings"
	"sync"
)

// TCPListener the tcp listener of multinet
type TCPListener struct {
	sync.Mutex

	listener *net.TCPListener

	ifListen   bool
	tcpChannel chan *TCPConn
	errChannel chan error

	groupTCPConn map[int]*groupTCPConn
}

// ListenTCP create the TCPListener of multinet
func ListenTCP(netStr string, laddr *net.TCPAddr) (*TCPListener, error) {
	tcpListener, err := net.ListenTCP(netStr, laddr)
	if err != nil {
		return nil, err
	}

	listener := new(TCPListener)
	listener.tcpChannel = make(chan *TCPConn, 1024)
	listener.errChannel = make(chan error, 1024)
	listener.listener = tcpListener
	listener.groupTCPConn = make(map[int]*groupTCPConn, 100)

	return listener, nil
}

// AcceptTCP get the virtual tcpconn of multinet
func (tl *TCPListener) AcceptTCP() (tcpConn *TCPConn, err error) {

	go tl.acceptTCP()
	select {
	case tcpConn = <-tl.tcpChannel:
	case err = <-tl.errChannel:
	}
	return
}

// acceptTCP accept new tcp, return when errors,
// but create new virtual tcp connection when new syncID data read
func (tl *TCPListener) acceptTCP() {
	tl.Lock()
	if tl.ifListen {
		tl.Unlock()
		return
	}
	tl.ifListen = true
	defer func() {
		tl.ifListen = false
	}()
	tl.Unlock()

	data := make([]byte, 1024)
	groupID := 0
	for {
		conn, err := tl.listener.AcceptTCP()
		if err != nil {
			tl.errChannel <- err
			return
		}

		count, err := conn.Read(data)
		if err != nil {
			conn.Close()
			tl.errChannel <- err
			return
		}

		dataStr := string(data[:count])
		clientID := getUniqueID()

		//new group
		if strings.Compare(dataStr, getGroupID) == 0 {
			groupID = getUniqueID()
			_, err = conn.Write([]byte(combineData(groupID, clientID)))
			if err != nil {
				conn.Close()
				tl.errChannel <- err
				return
			}

			groupTCPConn := newGroupTCPConn(groupID, "", nil, nil, tl)
			groupTCPConn.addRealConn(clientID, conn, "")

			tl.groupTCPConn[groupID] = groupTCPConn
			continue
		}

		//new group member
		if groupID, err = strconv.Atoi(dataStr); err == nil {
			_, err = conn.Write([]byte(combineData(groupID, clientID)))
			if err != nil {
				conn.Close()
				tl.errChannel <- err
				return
			}
			tl.groupTCPConn[groupID].addRealConn(clientID, conn, "")
		} else {
			conn.Close()
		}

	}
}

// Close the TCPListener
func (tl *TCPListener) Close() error {
	return tl.listener.Close()
}
