package multinet

import (
	"errors"
	"net"
	"strconv"
	"sync"
)

// realTCPConn the real tcp connection class
type realTCPConn struct {
	gtc                 *groupTCPConn
	clientID            int
	conn                *net.TCPConn
	canRead             bool
	canWrite            bool
	writeControlChannel chan bool
}

// newRealTCPConn create real tcp connection
func newRealTCPConn(gtc *groupTCPConn, clientID int, conn *net.TCPConn) *realTCPConn {
	return &realTCPConn{
		gtc:                 gtc,
		clientID:            clientID,
		conn:                conn,
		canRead:             true,
		canWrite:            true,
		writeControlChannel: make(chan bool),
	}
}

// read the data
func (rtc *realTCPConn) read(tmpData []byte) {
	data := make([]byte, 1024)
	nowDataLen := 0
	var n int
	var err error
	for {
		n, err = rtc.conn.Read(data)
		if err != nil {
			rtc.canRead = false
			rtc.gtc.delRealConn(rtc.clientID)
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
					rtc.canRead = false
					rtc.gtc.delRealConn(rtc.clientID)
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
			rtc.canRead = false
			rtc.gtc.delRealConn(rtc.clientID)
			return
		}

		tmpData = tmpData[nowDataLen:]
		nowDataLen = 0

		if pd.GroupID != rtc.gtc.groupID {
			continue
		}
		if tcpConn := rtc.gtc.getTCPConnBySyncID(pd.SyncID); tcpConn != nil {
			tcpConn.readChannel <- pd
		} else {
			putPackageData(pd)
		}

		if len(tmpData) != 0 {
			goto DealWithTmpData
		}
	}
}

// write the data
func (rtc *realTCPConn) write() {
	var err error
	for {
		select {
		case <-rtc.writeControlChannel:
			return
		case pd := <-rtc.gtc.writeChannel:
			data := pd.Encode()
			length := strconv.Itoa(len(data) + 4)
			for len(length) < 4 {
				length = "0" + length
			}
			data = append([]byte(length), data...)
			_, err = rtc.conn.Write(data)
			if err != nil {
				rtc.gtc.writeChannel <- pd
				rtc.canWrite = false
				return
			}
			putPackageData(pd)
		}
	}
}

// groupTCPConn is the core struct to control the tcp connections
type groupTCPConn struct {
	sync.Mutex
	groupID int
	cap     int

	netStr string
	laddr  *net.TCPAddr
	raddr  *net.TCPAddr

	errorChannel   chan error
	writeChannel   chan *packageData
	tcpConn        map[int]*realTCPConn
	virtualTCPConn map[int]*TCPConn

	listener *TCPListener
}

// newGroupTCPConn create the groupTCPConn
func newGroupTCPConn(groupID int, netStr string, laddr, raddr *net.TCPAddr, listener *TCPListener) *groupTCPConn {
	gtconn := new(groupTCPConn)
	gtconn.groupID = groupID

	gtconn.netStr = netStr
	gtconn.laddr = laddr
	gtconn.raddr = raddr

	gtconn.errorChannel = make(chan error, 1024)
	gtconn.writeChannel = make(chan *packageData, 1024)
	gtconn.tcpConn = make(map[int]*realTCPConn, 100)
	gtconn.virtualTCPConn = make(map[int]*TCPConn, 100)

	gtconn.listener = listener
	return gtconn
}

// addRealConn add real tcp connection to add group after dialing
func (gtc *groupTCPConn) addRealConn(clientID int, conn *net.TCPConn, tmpData string) {
	gtc.Lock()
	gtc.cap++
	rtc := newRealTCPConn(gtc, clientID, conn)
	gtc.tcpConn[clientID] = rtc
	go rtc.read([]byte(tmpData))
	go rtc.write()
	gtc.Unlock()
	return
}

// delRealConn delete real tcp connections when errors
func (gtc *groupTCPConn) delRealConn(clientID int) {
	gtc.Lock()
	rtc := gtc.tcpConn[clientID]
	if !rtc.canRead {
		if rtc.canWrite {
			rtc.writeControlChannel <- true
		}
		gtc.cap--
		delete(gtc.tcpConn, clientID)
	}
	gtc.Unlock()
}

// getTCPConnBySyncID create new virtual tcp connection when groupTCPCon get new syncID
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

// getTCPConn get virtual tcp connection
func (gtc *groupTCPConn) getTCPConn() (tcpConn *TCPConn) {
	gtc.Lock()
	tcpConn = newTCPConn(gtc, 0)
	gtc.virtualTCPConn[tcpConn.syncID] = tcpConn
	gtc.Unlock()
	return
}

// deleteTCPConn delete virtual tcp connection
func (gtc *groupTCPConn) deleteTCPConn(syncID int) {
	gtc.Lock()
	delete(gtc.virtualTCPConn, syncID)
	gtc.Unlock()
	return
}

// dial and create new real tcp connection
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
		gtc.addRealConn(cid, conn, tmpData)
	} else {
		conn.Close()
		return errors.New("verify group id fail")
	}

	return nil
}

// Close the groupTCPConn
func (gtc *groupTCPConn) Close() error {
	for _, realtcpconn := range gtc.tcpConn {
		if err := realtcpconn.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// globalMapGroupTCPConn the map to store the groupTCPConn
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
