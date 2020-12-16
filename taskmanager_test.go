package main

import "testing"

func TestInquiry(t *testing.T) {
	ttt := tasks{}
	ttt.Init(1)
	defer ttt.Uninit()
	ttt.ReadInquiry()

}
