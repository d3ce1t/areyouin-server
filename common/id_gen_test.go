package common

import (
	"testing"
)

func TestIDGenerator(t *testing.T) {

	hash := make(map[uint64]bool)
	idgen := NewIDGen(1)

	for i := 0; i < 40000; i++ {
		newId := idgen.GenerateID()

		if _, ok := hash[newId]; ok {
			t.Fatal("The generated ID", newId, "already exists on iteration", i)
		}

		hash[newId] = true
	}
}
