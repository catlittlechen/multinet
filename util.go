package multinet

import (
	"strconv"
	"strings"
	"sync"
)

func splitData(data string) (gid, cid int, err error) {
	arr := strings.Split(data, "&")
	gid, err = strconv.Atoi(arr[0])
	if err != nil {
		return
	}
	cid, err = strconv.Atoi(arr[1])
	return
}

type uniqueID struct {
	sync.Mutex
	id int
}

var globalUniqueID = new(uniqueID)

func getUniqueID() (id int) {
	globalUniqueID.Lock()
	globalUniqueID.id++
	id = globalUniqueID.id
	globalUniqueID.Unlock()
	return
}
