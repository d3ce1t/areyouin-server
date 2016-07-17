package idgen

import (
	"testing"
)

func TestIDGenerator(t *testing.T) {

	hash := make(map[uint64]bool)
	idgen := newIDGen(1)

	for i := 0; i < 40000; i++ {
		newId := idgen.generateID()

		if _, ok := hash[newId]; ok {
			t.Fatal("The generated ID", newId, "already exists on iteration", i)
		}

		hash[newId] = true
	}
}
