package multinet

import ()

type packageData struct {
	GroupID int    `json:"gid"`
	SynID   int    `json:"sid"`
	Data    string `json:"data"`
}

func NewPackageData(groupID int, b []byte) (d packageData) {
	d = packageData{
		GroupID: groupID,
		SynID:   getSynID(),
		Data:    string(b),
	}
	return
}
