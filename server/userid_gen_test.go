package main

import (
	"testing"
)

func TestUIDGenerator(t *testing.T) {

	hash := make(map[uint64]bool)
	uidgen := NewUIDGen(2)

	for i := 0; i < 10000; i++ {
		newId := uidgen.GenerateID()

		if _, ok := hash[newId]; ok {
			t.Fatal("The generated ID", newId, "already exists on iteration", i)
		}

		hash[newId] = true
	}
}
