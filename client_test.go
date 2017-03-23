package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCanHandlePass(t *testing.T) {
	c := new(Client)
	c.handleMessage("PASS test;oauth:123test")

	assert.Equal(t, "oauth:123test", c.bot.pass)

	c.handleMessage("NICK Nuuls")
	assert.Equal(t, "nuuls", c.bot.nick)
}
