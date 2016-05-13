package multinet

import (
	"strconv"
	"strings"
	"sync"
)

type packageData struct {
	GroupID int
	SyncID  int
	Data    string
}

func NewPackageData(groupID, syncID int, b []byte) (d *packageData) {
	d = getPackageData()
	d.GroupID = groupID
	d.SyncID = syncID
	d.Data = string(b)
	return
}

func (self *packageData) Encode() []byte {
	return []byte(strconv.Itoa(self.GroupID) + "&" + strconv.Itoa(self.SyncID) + "&" + self.Data)
}

func (self *packageData) Decode(data []byte) (err error) {
	array := strings.SplitN(string(data), "&", 3)
	self.GroupID, err = strconv.Atoi(array[0])
	if err != nil {
		return err
	}
	self.SyncID, _ = strconv.Atoi(array[1])
	if err != nil {
		return err
	}
	self.Data = array[2]
	return
}

type uniqueID struct {
	sync.Mutex
	id int
}

var globalUniqueID = new(uniqueID)

func getUniqueID() int {
	globalUniqueID.Lock()
	defer globalUniqueID.Unlock()
	globalUniqueID.id++
	return globalUniqueID.id
}
