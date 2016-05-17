package multinet

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type groupTCPConn struct {
	sync.Mutex
	groupID int
	cap     int

	netStr string
	laddr  *net.TCPAddr
	raddr  *net.TCPAddr

	errorChannel   chan error
	writeChannel   chan *packageData
	tcpConn        map[int]*net.TCPConn
	virtualTCPConn map[int]*TCPConn

	listener *TCPListener
}

func newGroupTCPConn(groupID int, netStr string, laddr, raddr *net.TCPAddr, listener *TCPListener) *groupTCPConn {
	gtconn := new(groupTCPConn)
	gtconn.groupID = groupID

	gtconn.netStr = netStr
	gtconn.laddr = laddr
	gtconn.raddr = raddr

	gtconn.errorChannel = make(chan error, 1024)
	gtconn.writeChannel = make(chan *packageData, 1024)
	gtconn.tcpConn = make(map[int]*net.TCPConn, 100)
	gtconn.virtualTCPConn = make(map[int]*TCPConn, 100)

	gtconn.listener = listener
	return gtconn
}

func (gtc *groupTCPConn) addConn(clientID int, conn *net.TCPConn, tmpData string) {
	gtc.Lock()
	gtc.cap++
	gtc.tcpConn[clientID] = conn
	go gtc.read(clientID, conn, []byte(tmpData))
	go gtc.write(clientID, conn)
	gtc.Unlock()
	return
}

func (gtc *groupTCPConn) read(clientID int, conn *net.TCPConn, tmpData []byte) {
	data := make([]byte, 1024)
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

		if pd.GroupID != gtc.groupID {
			continue
		}
		if tcpConn := gtc.getTCPConnBySyncID(pd.SyncID); tcpConn != nil {
			tcpConn.readChannel <- pd
		} else {
			putPackageData(pd)
		}

		if len(tmpData) != 0 {
			goto DealWithTmpData
		}
	}
}

func (gtc *groupTCPConn) getTCPConnBySyncID(syncID int) (tcpConn *TCPConn) {
	gtc.Lock()
	defer gtc.Unlock()

	ok := false
	if tcpConn, ok = gtc.virtualTCPConn[syncID]; !ok {
		if gtc.listener != nil {
			tcpConn = newTCPConn(gtc, syncID)
			gtc.listener.tcpChannel <- tcpConn
			gtc.virtualTCPConn[tcpConn.syncID] = tcpConn
		}
	}

	return
}

func (gtc *groupTCPConn) write(clientID int, conn *net.TCPConn) {
	var err error
	for {
		pd := <-gtc.writeChannel
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
		putPackageData(pd)
	}
}

func (gtc *groupTCPConn) Close() error {
	for _, tcpconn := range gtc.tcpConn {
		if err := tcpconn.Close(); err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (gtc *groupTCPConn) getTCPConn() (tcpConn *TCPConn) {
	gtc.Lock()
	tcpConn = newTCPConn(gtc, 0)
	gtc.virtualTCPConn[tcpConn.syncID] = tcpConn
	gtc.Unlock()
	return
}

func (gtc *groupTCPConn) deleteTCPConn(syncID int) {
	gtc.Lock()
	delete(gtc.virtualTCPConn, syncID)
	gtc.Unlock()
	return
}

func (gtc *groupTCPConn) dial() error {
	if gtc.cap >= maxTCPCount {
		return nil
	}

	conn, err := net.DialTCP(gtc.netStr, gtc.laddr, gtc.raddr)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(strconv.Itoa(gtc.groupID)))
	if err != nil {
		conn.Close()
		return err
	}

	data := make([]byte, 100)
	count, err := conn.Read(data)
	if err != nil {
		conn.Close()
		return err
	}
	gid, cid, tmpData, err := splitData(string(data[:count]))
	if err != nil {
		conn.Close()
		return err
	}

	if gtc.groupID == gid {
		gtc.addConn(cid, conn, tmpData)
	} else {
		conn.Close()
		return errors.New("verify group id fail")
	}

	return nil
}

var globalMapGroupTCPConn = make(map[string]*groupTCPConn, 100)

func getGroupConn(netStr string, laddr, raddr *net.TCPAddr) *groupTCPConn {
	key := netStr + "&" + laddr.String() + "&" + raddr.String()
	return globalMapGroupTCPConn[key]
}

func setGroupConn(netStr string, laddr, raddr *net.TCPAddr, gtc *groupTCPConn) {
	key := netStr + "&" + laddr.String() + "&" + raddr.String()
	globalMapGroupTCPConn[key] = gtc
	return
}
