package main

import (
	"sync"
	"testing"
)

// not working at the moment, output of the powershell with this given file.
// GolandProjects\awesomeProject1> go test -bench=.
// ok      awesomeProject1 0.247s [no tests to run]

func Benchmark_takeFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		urlFlowSender := make(chan string, workersCnt)
		var readFileWg sync.WaitGroup

		// Start the benchmark timer
		b.ResetTimer()
		readFileWg.Add(1)
		go takeFile(&readFileWg, InputFileAddress, urlFlowSender)
		readFileWg.Wait()

		// Drain the channel to ensure it's empty
		for range urlFlowSender {
		}
	}
}
