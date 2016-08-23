package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"peeple/areyouin/cqldao"
	"peeple/areyouin/utils"
	"sync"
	"time"

	hist "github.com/uniplot/histogram"
)

type testHandler func(testNumber int) (time.Duration, error)

var availableTests = map[string]testHandler{
	"create_event": testCreateEvent,
}

var dbSession *cqldao.GocqlSession
var eventDAO *cqldao.EventDAO

type executionStats struct {
	min         time.Duration
	max         time.Duration
	avg         time.Duration
	cdur        time.Duration
	ops         int
	samples     []time.Duration
	numErrors   int
	numTimes    int
	startOffset int
	endOffset   int
}

func (s executionStats) String() string {

	minStr := s.min.String()
	maxStr := s.max.String()
	avgStr := s.avg.String()
	opsStr := fmt.Sprintf("%v", s.ops)

	if s.min == math.MaxInt64 {
		minStr = "-"
	}

	if s.max == 0 {
		maxStr = "-"
	}

	if s.avg == 0 {
		avgStr = "-"
	}

	if s.ops == 0 {
		opsStr = "-"
	}

	return fmt.Sprintf("min: %13v | max: %13v | avg: %13v | avg.ops: %4v | samples: %5v | errors: %4v | total: %5v | start: %5v | end: %5v",
		minStr, maxStr, avgStr, opsStr, len(s.samples), s.numErrors, s.numTimes, s.startOffset, s.endOffset)
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
	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}

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
	dbSession.Connect()

	if !dbSession.IsValid() || dbSession.Closed() {
		os.Exit(2)
	}

	// Init DAOs
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

	statsSlice := make([]*executionStats, numWorkers)

	// Distribute work between workers

	startTime := time.Now()
	startIndex := 0

	for i := 0; i < numWorkers; i++ {

		wg.Add(1)

		availableThreads := numWorkers - i
		workSize := int(math.Trunc(float64(totalWork) / float64(availableThreads)))
		if totalWork%availableThreads > 0 {
			workSize++
		}

		// Do things
		go func(workerId int, startOffset int, endOffset int) {
			stats := executeTestInWorker(t, workerId, startOffset, endOffset)
			statsSlice[workerId] = stats
			wg.Done()
		}(i, startIndex, startIndex+workSize-1)

		// Decrease remaining work
		totalWork -= workSize
		startIndex += workSize
	}

	wg.Wait()

	duration := time.Now().Sub(startTime)

	// Print local stats
	/*for workerId, stats := range statsSlice {
		fmt.Printf("Worker: %3v | %v\n", workerId, stats)
	}*/

	// Analyse data
	globalStats := computeGlobalStats(statsSlice, duration)
	histogram := computeHistogram(globalStats)

	// Show histogram
	maxWidth := 10
	err := hist.Fprint(os.Stdout, histogram, hist.Linear(maxWidth))
	if err != nil {
		log.Printf("Histogram: %v", err)
	}

	// Shpw global stats
	fmt.Printf("Global: %v | %v\n", "---", globalStats)
	fmt.Printf("Bench total time: %v\n", duration)
}

func executeTestInWorker(test testHandler, workerId int, startOffset int, endOffset int) *executionStats {

	stats := &executionStats{
		numTimes:    endOffset - startOffset + 1,
		startOffset: startOffset,
		endOffset:   endOffset,
		min:         time.Duration(math.MaxInt64),
		max:         0,
	}

	for i := 0; i < stats.numTimes; i++ {

		duration, err := test(startOffset + i)

		if err == nil {
			stats.cdur += duration
			stats.samples = append(stats.samples, duration)
			stats.min = utils.MinDuration(stats.min, duration)
			stats.max = utils.MaxDuration(stats.max, duration)
		} else {
			stats.numErrors++
		}
	}

	numSamples := len(stats.samples)

	if numSamples > 0 {
		stats.ops = int(float64(numSamples) / stats.cdur.Seconds())
		stats.avg = time.Duration(int64(float64(stats.cdur) / float64(numSamples)))
	}

	return stats
}

func computeGlobalStats(statsSlice []*executionStats, globalDuration time.Duration) *executionStats {

	globalStats := &executionStats{
		startOffset: math.MaxInt64,
		endOffset:   0,
		min:         time.Duration(math.MaxInt64),
		max:         0,
	}

	for _, stats := range statsSlice {
		globalStats.min = utils.MinDuration(globalStats.min, stats.min)
		globalStats.max = utils.MaxDuration(globalStats.max, stats.max)
		globalStats.cdur += stats.cdur
		globalStats.samples = append(globalStats.samples, stats.samples...)
		globalStats.numErrors += stats.numErrors
		globalStats.numTimes += stats.numTimes
		globalStats.startOffset = utils.MinInt(globalStats.startOffset, stats.startOffset)
		globalStats.endOffset = utils.MaxInt(globalStats.endOffset, stats.endOffset)
	}

	numSamples := len(globalStats.samples)

	if numSamples > 0 {
		globalStats.ops = int(float64(numSamples) / globalDuration.Seconds())
		globalStats.avg = time.Duration(int64(float64(globalStats.cdur) / float64(numSamples)))
	}

	return globalStats
}

func computeHistogram(globalStats *executionStats) hist.Histogram {

	values := make([]float64, 0, len(globalStats.samples))

	for _, duration := range globalStats.samples {
		ms := math.Ceil(float64(duration.Nanoseconds()) / float64(1e6))
		values = append(values, ms)
	}

	numBeans := int(globalStats.max / globalStats.min)
	histogram := hist.Hist(numBeans, values)

	return histogram
}
