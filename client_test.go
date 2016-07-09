package main

import (
	"testing"
)

func TestCanHandlePass(t *testing.T) {
	c := new(Client)
	c.handleMessage("PASS test;oauth:123test")
	if c.bot.pass != "oauth:123test" {
		t.Fatal("pass doesn't match")
	}
	c.handleMessage("NICK Nuuls")
	if c.bot.nick != "nuuls" {
		t.Fatal("nick doesn't match")
	}
}
