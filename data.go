package multinet

import (
	"sync"
)

type packageData struct {
	GroupID int    `json:"gid"`
	SynID   int    `json:"sid"`
	Data    string `json:"data"`
}

func NewPackageData(groupID, syncID int, b []byte) (d packageData) {
	d = packageData{
		GroupID: groupID,
		SynID:   syncID,
		Data:    string(b),
	}
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
