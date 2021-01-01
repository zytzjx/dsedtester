package main

import (
	"strings"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
func TestInquiry(t *testing.T) {
	CreateRedisPool(GetLabelsCnt())
	ttt := tasks{}
	ttt.Init(1)
	defer ttt.Uninit()
	ttt.ReadInquiry()

}

func TestStringCompare(t *testing.T) {
	aa := "\tEnabled Supported:"
	ss := strings.Trim(aa, " \t")
	assertEqual(t, ss, "Enabled Supported:")
}

func TestIdentify(t *testing.T) {
	CreateRedisPool(GetLabelsCnt())
	ttt := tasks{}
	ttt.Init(1)
	defer ttt.Uninit()
	ttt.DriverIdentifyData()

}

func TestTestIDCheck(t *testing.T) {
	CreateRedisPool(GetLabelsCnt())
	ttt := tasks{}
	ttt.Init(1)
	defer ttt.Uninit()
	ttt.TestIDCheck()

}
