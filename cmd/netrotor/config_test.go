package main

import (
	"testing"
)

func TestConfig(t *testing.T) {
	// t.Fatal("not implemented")

	_, err := ReadConfig("test-fixtures/multirotorconf.json")

	if err != nil {
		t.Fatalf("readConfig() returned unexpected error: %v", err)
	}
}
