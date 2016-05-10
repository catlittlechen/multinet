package multinet

import (
	"strconv"
	"strings"
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
