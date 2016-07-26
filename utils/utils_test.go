package utils

import (
	"log"
	"testing"
	"time"
)

func TestUnixMillisToTime(t *testing.T) {
	current_time := time.Now()
	log.Println("Current", current_time)
	time := UnixMillisToTime(TimeToMillis(current_time))
	log.Println("Retrieve", time)
}
