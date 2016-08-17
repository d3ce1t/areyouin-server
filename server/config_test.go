package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	_, err := LoadConfigFromFile("config.example")
	if err != nil {
		t.Fatal(err)
	}
}
