package idgen

import (
	"time"
)

const epoch int64 = 1446336000000 // Milliseconds since 1 Nov 2015 00:00

var default_generator *IDGen
var id_gen_ch chan uint64

func init() {

	id_gen_ch = make(chan uint64)

	// Creates a generator with group id = 1
	default_generator := newIDGen(1)

	go func() {
		for {
			newId := default_generator.generateID()
			id_gen_ch <- newId
		}
	}()
}

func NewID() int64 {
	return int64(<-id_gen_ch)
}

func newIDGen(id uint16) *IDGen {
	return &IDGen{id: id, auto_increment: 0}
}

type IDGen struct {
	id             uint16 // 12 bits used (4096 different values)
	auto_increment uint16 // 10 bits used (1024 different values)
	last_time      time.Time
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
func (uid *IDGen) generateID() uint64 {

	curr_time := time.Now().UTC()

	if uid.auto_increment == 0 {
		diff_time := curr_time.Sub(uid.last_time)
		for diff_time <= 2*time.Millisecond {
			time.Sleep(2*time.Millisecond - diff_time)
			curr_time = time.Now().UTC()
			diff_time = curr_time.Sub(uid.last_time)
		}
		uid.last_time = curr_time
	}

	curr_time_ms := curr_time.UnixNano() / int64(time.Millisecond)
	curr_time_ms -= epoch

	newId := uint64(curr_time_ms) << (64 - 42) // 139 years of IDs
	newId |= uint64(uid.id) << (64 - 42 - 12)

	auto_inc := uid.auto_increment
	newId |= uint64(auto_inc)

	uid.auto_increment = (auto_inc + 1) % 1024
	return newId
}

// TODO: Load state from Cassandra DB
func (uid *IDGen) LoadState() {

}

// TODO: Save state in Cassandra DB
func (uid *IDGen) SaveState() {

}
