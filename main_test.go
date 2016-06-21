package main

import (
	"reflect"
	"testing"
)

func TestCanInitLogger(t *testing.T) {
	log := initLogger()
	if reflect.TypeOf(log).String() != "logging.Logger" {
		t.Fatal("logger invalid type")
	}
}
