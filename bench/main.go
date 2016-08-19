package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"peeple/areyouin/cqldao"
	"peeple/areyouin/utils"
	"sync"
	"time"
)

type testHandler func() (time.Duration, error)

var availableTests = map[string]testHandler{
	"create_event": testCreateEvent,
}

var dbSession *cqldao.GocqlSession
var eventDAO *cqldao.EventDAO

type executionStats struct {
	min        time.Duration
	max        time.Duration
	avg        time.Duration
	cdur       time.Duration
	ops        int
	numSamples int
	numErrors  int
	numTimes   int
}

func (s executionStats) String() string {
	return fmt.Sprintf("min: %13v | max: %13v | avg: %13v | avg.ops: %4v | samples: %5v | errors: %4v | total: %5v",
		s.min, s.max, s.avg, s.ops, s.numSamples, s.numErrors, s.numTimes)
}

// bench -h 127.0.0.1 -k areyouin -t create_event -n 20 -c 3

func showError(errStr string) {
	fmt.Printf("\n\tError: %v\n\n", errStr)
	fmt.Printf("\t%v --help for usage information\n\n", os.Args[0])
}

func main() {

	// Init flags

	var host string
	var keyspace string
	var cqlVersion int
	var test string
	var numTimes int
	var numThreads int

	flag.StringVar(&host, "h", "localhost", "IP address of Cassandra")
	flag.StringVar(&keyspace, "k", "", "Keyspace")
	flag.IntVar(&cqlVersion, "cql-version", 2, "CQL version")
	flag.StringVar(&test, "t", "", "Test name")
	flag.IntVar(&numTimes, "n", 1, "Times test will be executed")
	flag.IntVar(&numThreads, "c", 1, "Number of concurrent workers")

	flag.Parse()

	if keyspace == "" {
		showError("Keyspace name isn't set")
		return
	}

	if test == "" {
		showError("Test name isn't set")
		return
	}

	if cqlVersion < 2 || cqlVersion > 4 {
		showError("CQL version must be between 2 and 4")
		return
	}

	// Connect to database

	dbSession = cqldao.NewSession(keyspace, cqlVersion, host)
	eventDAO = cqldao.NewEventDAO(dbSession).(*cqldao.EventDAO)

	// Select

	if t, ok := availableTests[test]; ok {

		// Execute test

		executeTest(t, numTimes, numThreads)

	} else {
		showError("Selected test doesn't exist")
	}

}

func executeTest(t testHandler, numTimes int, numWorkers int) {

	var wg sync.WaitGroup
	totalWork := numTimes

	statsSlice := make([]executionStats, numWorkers)

	// Distribute work between workers

	startTime := time.Now()

	for i := 0; i < numWorkers; i++ {

		wg.Add(1)

		availableThreads := numWorkers - i
		workSize := int(math.Trunc(float64(totalWork) / float64(availableThreads)))
		if totalWork%availableThreads > 0 {
			workSize++
		}

		// Do things
		go func(workerId int) {
			stats := executeTestInWorker(t, workSize)
			statsSlice[workerId] = stats
			// Print individual stats
			fmt.Printf("Worker: %3v | %v\n", workerId, stats)
			wg.Done()
		}(i)

		// Decrease remaining work
		totalWork -= workSize
	}

	wg.Wait()

	duration := time.Now().Sub(startTime)

	// Print global stats
	globalStats := computeGlobalStats(statsSlice, duration)
	fmt.Printf("Global: %v | %v\n", "---", globalStats)
}

func executeTestInWorker(test testHandler, numTimes int) executionStats {

	var min int64 = math.MaxInt64
	var max int64
	var sumDur time.Duration

	numSamples := 0
	numErrors := 0

	for i := 0; i < numTimes; i++ {

		duration, err := test()

		if err == nil {
			sumDur += duration
			durInt64 := int64(duration)
			min = utils.MinInt64(min, durInt64)
			max = utils.MaxInt64(max, durInt64)
			numSamples++
		} else {
			numErrors++
		}
	}

	opsPerSecond := float64(numSamples) / sumDur.Seconds()
	avg := int64(float64(sumDur) / float64(numSamples))

	return executionStats{
		min:        time.Duration(min),
		max:        time.Duration(max),
		avg:        time.Duration(avg),
		cdur:       sumDur,
		ops:        int(opsPerSecond),
		numSamples: numSamples,
		numErrors:  numErrors,
		numTimes:   numTimes,
	}
}

func computeGlobalStats(statsSlice []executionStats, globalDuration time.Duration) executionStats {

	var min int64 = math.MaxInt64
	var max int64
	var cdur int64
	var numSamples int
	var numErrors int
	var numTimes int

	for _, stats := range statsSlice {
		min = utils.MinInt64(min, int64(stats.min))
		max = utils.MaxInt64(max, int64(stats.max))
		cdur += int64(stats.cdur)
		numSamples += stats.numSamples
		numErrors += stats.numErrors
		numTimes += stats.numTimes
	}

	opsPerSecond := float64(numSamples) / globalDuration.Seconds()
	avg := int64(float64(cdur) / float64(numSamples))

	return executionStats{
		min:        time.Duration(min),
		max:        time.Duration(max),
		avg:        time.Duration(avg),
		cdur:       time.Duration(cdur),
		ops:        int(opsPerSecond),
		numSamples: numSamples,
		numErrors:  numErrors,
		numTimes:   numTimes,
	}
}
