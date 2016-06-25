package main

import (
	"reflect"
	"testing"
)

func TestCanCreateNewConnection(t *testing.T) {
	ct := newConnection("gempir", "password")
	if reflect.TypeOf(ct).String() != "main.Connection" {
		t.Fatal("new connection of wrong type", reflect.TypeOf(ct).String())
	}
}
