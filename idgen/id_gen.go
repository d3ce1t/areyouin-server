package idgen

import (
	"time"
)

const epoch int64 = 1446336000000 // Milliseconds since 1 Nov 2015 00:00

var defaultGenerator *IDGen
var idGenCh chan uint64

func init() {

	idGenCh = make(chan uint64)

	// Creates a generator with group id = 1
	defaultGenerator := newIDGen(1)
	var newID uint64

	go func() {
		for {
			newID = defaultGenerator.generateID()
			idGenCh <- newID
		}
	}()
}

func NewID() int64 {
	return int64(<-idGenCh)
}

func newIDGen(id uint16) *IDGen {
	return &IDGen{id: id % 4096, autoIncrement: 0}
}

type IDGen struct {
	id             uint16 // 12 bits used (4096 different values)
	autoIncrement  uint16 // 10 bits used (1024 different values)
	lastTime       time.Time
	sleepCounter   uint64
	resleepCounter uint64
}

/*
  Generates an ID of 64 bits where the first most significant 42 bits
  are the current time in millis since epoch (1 nov 2015 00:00:00), next
  12 bits are the ID of this generator (to avoid collisions) and the
  last 10 bits are an auto_increment number.

  Each generator that is executed in a different process or thread must
  have a different ID in order to avoid collisions.

  Because of the auto_increment number, a single generator can generate
  up to 1024 different IDs per millisecond. If this limit is exceeded
  IDs will repeat what could cause undefined behaviour to whatever
  who uses this generator.
*/
func (uid *IDGen) generateID() uint64 {

	// Golang time package takes into account summer time changes. So
	// timestamp doesn't repeat in those cases.
	// https://play.golang.org/p/zZcI9QD_Zm
	currTime := time.Now().UTC()

	// Auto increment logic
	diffTime := currTime.Sub(uid.lastTime)
	if diffTime < 1*time.Millisecond {

		uid.autoIncrement = (uid.autoIncrement + 1) % 1024

		if uid.autoIncrement == 0 {
			// AutoIncrement values exhausted for this millisecond
			// Move to next millisecond after lastTime
			uid.waitTillNextMillisecond()
			currTime = time.Now().UTC()
		}

	} else {
		uid.autoIncrement = 0
	}

	currTimeMs := currTime.UnixNano() / int64(time.Millisecond)
	currTimeMs -= epoch

	newID := uint64(currTimeMs) << (64 - 42) // 139 years of IDs
	newID |= uint64(uid.id) << (64 - 42 - 12)
	newID |= uint64(uid.autoIncrement)

	uid.lastTime = currTime

	return newID
}

// waitTillNextMillisecond waits until current time is 1ms after last time a ID was
// generated. Because of this, this implementation is not affected by leap seconds.
// In the curse of a leap second, in the worst case (lastTime 23:59:60.999,
// timestamp x.999), a call to this function at 00:00:00.000 would be seen as the
// repeated second 23:59:60, with same timestamp. In this case, the caller would
// have to wait 1.001 second, instead of only 1 ms.
func (uid *IDGen) waitTillNextMillisecond() {

	uid.sleepCounter++
	currTime := time.Now().UTC()

	for currTime.Sub(uid.lastTime) < 1*time.Millisecond {
		uid.resleepCounter++
		time.Sleep(1 * time.Millisecond)
		currTime = time.Now().UTC()
	}
}

// TODO: Load state from Cassandra DB
func (uid *IDGen) LoadState() {

}

// TODO: Save state in Cassandra DB
func (uid *IDGen) SaveState() {

}
