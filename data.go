package multinet

import (
	"strconv"
	"strings"
)

type packageData struct {
	GroupID int
	SyncID  int
	Data    string
}

func newPackageData(groupID, syncID int, b []byte) (d *packageData) {
	d = getPackageData()
	d.GroupID = groupID
	d.SyncID = syncID
	d.Data = string(b)
	return
}

func (pd *packageData) Encode() []byte {
	return []byte(strconv.Itoa(pd.GroupID) + "&" + strconv.Itoa(pd.SyncID) + "&" + pd.Data)
}

func (pd *packageData) Decode(data []byte) (err error) {
	array := strings.SplitN(string(data), "&", 3)
	pd.GroupID, err = strconv.Atoi(array[0])
	if err != nil {
		return err
	}
	pd.SyncID, err = strconv.Atoi(array[1])
	if err != nil {
		return err
	}
	pd.Data = array[2]
	return
}
