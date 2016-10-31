package idgen

import (
	"testing"
)

func TestIDGenerator(t *testing.T) {

	hash := make(map[uint64]bool)
	idgen := newIDGen(1)

	for i := 0; i < 2000000; i++ {

		newID := idgen.generateID()

		if _, ok := hash[newID]; ok {
			t.Fatal("The generated ID", newID, "already exists on iteration", i)
		}

		hash[newID] = true
	}

	t.Logf("Sleep Counter: %v, ReSleep Counter: %v\n", idgen.sleepCounter, idgen.resleepCounter-idgen.sleepCounter)
}
