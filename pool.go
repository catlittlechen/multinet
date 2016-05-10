package multinet

import (
	"sync"
)

var packageDataPool = &sync.Pool{
	New: func() interface{} {
		return new(packageData)
	},
}

func getPackageData() *packageData {
	return packageDataPool.Get().(*packageData)
}

func putPackageData(p *packageData) {
	packageDataPool.Put(p)
	return
}
