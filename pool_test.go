package multinet

import (
	"log"
	"testing"
)

func TestPool(t *testing.T) {
	p := getPackageData()
	log.Print(p)
	p.SynID = 100
	putPackageData(p)
	p2 := getPackageData()
	if p2.SynID != 100 {
		t.Fatalf("%+v\n", p2)
	}
	p3 := getPackageData()
	if p3.SynID == 100 {
		t.Fatalf("%+v\n", p3)
	}
	t.Log("ok")
}
