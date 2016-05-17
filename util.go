package multinet

import (
	"strconv"
	"strings"
	"sync"
)

func splitData(data string) (gid, cid int, tmpData string, err error) {
	arr := strings.SplitN(data, "&", 3)
	gid, err = strconv.Atoi(arr[0])
	if err != nil {
		return
	}
	cid, err = strconv.Atoi(arr[1])
	tmpData = arr[2]
	return
}

func combineData(gid, cid int) (data string) {
	data = strconv.Itoa(gid) + "&" + strconv.Itoa(cid) + "&"
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
