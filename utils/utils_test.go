package utils

import (
	"log"
	"testing"
	"time"
)

func TestUnixMillisToTime(t *testing.T) {
	currentTime := time.Now()
	log.Println("Current", currentTime)
	time := MillisToTimeUTC(TimeToMillis(currentTime))
	log.Println("Retrieve", time)
}
