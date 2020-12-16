package main

import "testing"

func TestHandlelogprogress(t *testing.T) {
	CreateRedisPool(5)
	line := `   1      1 0xff   0.159%   0.159% 00:00:05 00:00:05 10:38:01 00003141    93.18    93.18`
	label := 1
	handlelogprogress(label, line)
}
