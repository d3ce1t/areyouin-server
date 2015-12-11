package main

import (
	"time"
)

const epoch int64 = 1446336000000 // Milliseconds since 1 Nov 2015 00:00

func NewUIDGen(id uint16) *UIDGen {
	return &UIDGen{id: id, auto_increment: 0}
}

type UIDGen struct {
	id             uint16 // 12 bits used (4096 different values)
	auto_increment uint16 // 10 bits used (1024 different values)
}

/*
  Generates an ID of 64 bits where the first most significant 42 bits
  are the current time in millis since epoch (1 nov 2015 00:00), next
  12 bits are the ID of this generator (to avoid collisions) and the
  last 10 bits are an auto_increment number.

  Each generator that is executed in a different process or thread must
  have a different ID in order to avoid collisions.

  Because of the auto_increment number, a single generator can generate
  up to 1024 different IDs per millisecond
*/
func (uid *UIDGen) GenerateID() uint64 {
	curr_time := time.Now().UTC().UnixNano() / int64(time.Millisecond)
	curr_time -= epoch
	newId := uint64(curr_time) << (64 - 42)
	newId |= uint64(uid.id) << (64 - 42 - 12)
	auto_inc := uid.auto_increment % 1024
	newId |= uint64(auto_inc)
	uid.auto_increment = (auto_inc + 1) % 1024
	return newId
}

func (uid *UIDGen) LoadState() {

}

func (uid *UIDGen) SaveState() {

}
