package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanConnectToServer(t *testing.T) {
	server := NewServer(3000)

	assert.IsType(t, tcpServer{}, server)
}
