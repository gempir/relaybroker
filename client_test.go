package main

import (
	"testing"
)

func TestCanHandlePass(t *testing.T) {
	c := new(Client)
	c.handleMessage("PASS test;oauth:123test")
	if c.pass != "test;oauth:123test" {
		t.Fatal("pass doesn't match")
	}
	c.handleMessage("NICK Nuuls")
	if c.nick != "nuuls" {
		t.Fatal("nick doesn't match")
	}
}
