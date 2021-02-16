package main

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
func TestInquiry(t *testing.T) {
	// CreateRedisPool(GetLabelsCnt())
	// ttt := tasks{}
	// ttt.Init(1)
	// defer ttt.Uninit()
	// ttt.ReadInquiry()

	logs, err := os.OpenFile("logs/test.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	///dev/sde   /dev/sg5
	///dev/sdd   /dev/sg4
	task := &tasks{
		mu:        &sync.Mutex{},
		logfile:   logs,
		tasklist:  []string{},
		lastError: nil,
		label:     1,
		cmddict:   map[int]*exec.Cmd{},
		linuxName: "/dev/sde",
		sgName:    "/dev/sg5",
	}
	assertEqual(t, task.ReadInquiry(), nil)
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

func TestTestCryptoScramble(t *testing.T) {
	logs, err := os.OpenFile("logs/test.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	///dev/sde   /dev/sg5
	///dev/sdd   /dev/sg4
	task := &tasks{
		mu:        &sync.Mutex{},
		logfile:   logs,
		tasklist:  []string{},
		lastError: nil,
		label:     1,
		cmddict:   map[int]*exec.Cmd{},
		linuxName: "/dev/sde",
		sgName:    "/dev/sg5",
	}
	assertEqual(t, task.TestCryptoScramble(), nil)
}

func TestTestMaxAddress(t *testing.T) {
	logs, err := os.OpenFile("logs/test.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	///dev/sde   /dev/sg5
	///dev/sdd   /dev/sg4
	task := &tasks{
		mu:        &sync.Mutex{},
		logfile:   logs,
		tasklist:  []string{},
		lastError: nil,
		label:     1,
		cmddict:   map[int]*exec.Cmd{},
		linuxName: "/dev/sde",
		sgName:    "/dev/sg5",
	}
	assertEqual(t, task.TestMaxAddress(), nil)
}

func TestDriverReadInfo(t *testing.T) {
	logs, err := os.OpenFile("logs/test.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	///dev/sde   /dev/sg5
	///dev/sdd   /dev/sg4
	task := &tasks{
		mu:        &sync.Mutex{},
		logfile:   logs,
		tasklist:  []string{},
		lastError: nil,
		label:     1,
		cmddict:   map[int]*exec.Cmd{},
		linuxName: "/dev/sde",
		sgName:    "/dev/sg5",
	}
	assertEqual(t, task.DriverReadInfo(), nil)
}
