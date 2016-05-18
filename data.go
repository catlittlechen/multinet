package multinet

import (
	"errors"
	"strconv"
	"strings"
)

type packageData struct {
	GroupID int
	SyncID  int
	Code    int
	Data    string
}

func newPackageData(groupID, syncID, code int, b []byte) (d *packageData) {
	d = getPackageData()
	d.GroupID = groupID
	d.SyncID = syncID
	d.Code = code
	d.Data = string(b)
	return
}

func (pd *packageData) Encode() []byte {
	return []byte(strconv.Itoa(pd.GroupID) + "&" + strconv.Itoa(pd.SyncID) + "&" + strconv.Itoa(pd.Code) + "&" + pd.Data)
}

func (pd *packageData) Decode(data []byte) (err error) {
	array := strings.SplitN(string(data), "&", 4)
	if len(array) != 4 {
		return errors.New("data struct is wrong")
	}
	pd.GroupID, err = strconv.Atoi(array[0])
	if err != nil {
		return err
	}
	pd.SyncID, err = strconv.Atoi(array[1])
	if err != nil {
		return err
	}

	pd.Code, err = strconv.Atoi(array[2])
	if err != nil {
		return err
	}

	pd.Data = array[3]
	return
}
