package main

import "testing"

func TestCreateRedisPool(t *testing.T) {
	CreateRedisPool(5)
	ping(1)
}
