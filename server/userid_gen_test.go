package main

import (
	"fmt"
	"testing"
	"time"
)

func TestUIDGenerator(t *testing.T) {
	uidgen := NewUIDGen(2)
	for i := 0; i < 10; i++ {
		newId := uidgen.GenerateID()
		fmt.Println(newId, uidgen)
		time.Sleep(1 * time.Millisecond)
	}
}
