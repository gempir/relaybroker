package main

import (
	"testing"

	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
)

func TestCanInitLogger(t *testing.T) {
	log := initLogger(0)
	assert.IsType(t, logging.Logger{}, log)
}
