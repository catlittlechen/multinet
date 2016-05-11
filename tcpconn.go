package multinet

import (
	"errors"
	"net"
	"strconv"
)

type TCPConn struct {
	groupConn *GroupTCPConn
	syncID    int
}

func DialTCP(netStr string, laddr, raddr *net.TCPAddr) (*TCPConn, error) {

	// only first dial will work
	if tgc := getGroupConn(netStr, laddr, raddr); tgc != nil {
		return tgc.getTCPConn(), nil
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

	groupTCPConn := &GroupTCPConn{
		groupID:        gid,
		tcpConn:        make(map[int]*net.TCPConn),
		virtualTCPConn: make(map[int]*TCPConn),
	}

	groupTCPConn.addConn(cid, conn)

	for i := 1; i < TCPCount; i++ {

		conn, err = net.DialTCP(netStr, laddr, raddr)
		if err != nil {
			return nil, err
		}

		_, err = conn.Write([]byte(strconv.Itoa(groupTCPConn.groupID)))
		if err != nil {
			conn.Close()
			return nil, err
		}
		count, err = conn.Read(data)
		if err != nil {
			conn.Close()
			return nil, err
		}
		gid, cid, err = splitData(string(data[:count]))
		if err != nil {
			conn.Close()
			return nil, err
		}

		if groupTCPConn.groupID == gid {
			groupTCPConn.addConn(cid, conn)
		} else {
			conn.Close()
			return nil, errors.New("verify group id fail")
		}
	}

	setGroupConn(netStr, laddr, raddr, groupTCPConn)
	return groupTCPConn.getTCPConn(), nil
}

func (self *TCPConn) Close() {
	delete(self.groupConn.virtualTCPConn, self.syncID)
}

func (self *TCPConn) Read() {
}

func (self *TCPConn) Write() {
}
