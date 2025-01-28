package bench

import (
	"fmt"
	"testing"
)

func Benchmark(parallelism int, dicedbBenchFunc func(b *testing.B)) {
	nsPerOpChan := make(chan float64, parallelism)
	allocsPerOpChan := make(chan int64, parallelism)
	bytesPerOpChan := make(chan int64, parallelism)
	throughputChan := make(chan float64, parallelism)

	for i := 0; i < parallelism; i++ {
		go func() {
			results := testing.Benchmark(dicedbBenchFunc)
			nsPerOpChan <- float64(results.NsPerOp())
			allocsPerOpChan <- results.AllocsPerOp()
			bytesPerOpChan <- results.AllocedBytesPerOp()
			throughputChan <- float64(1e9) / float64(results.NsPerOp())
		}()
	}

	var totalNsPerOp, totalThroughput float64
	var totalAllocsPerOp, totalBytesPerOp int64

	for i := 0; i < parallelism; i++ {
		totalNsPerOp += <-nsPerOpChan
		totalAllocsPerOp += <-allocsPerOpChan
		totalBytesPerOp += <-bytesPerOpChan
		totalThroughput += <-throughputChan
	}

	avgNsPerOp := totalNsPerOp / float64(parallelism)
	avgAllocsPerOp := totalAllocsPerOp / int64(parallelism)
	avgBytesPerOp := totalBytesPerOp / int64(parallelism)

	fmt.Printf("parallelism: %d\n", parallelism)
	fmt.Printf("avg ns/op: %.2f\n", avgNsPerOp)
	fmt.Printf("avg allocs/op: %d\n", avgAllocsPerOp)
	fmt.Printf("avg bytes/op: %d\n", avgBytesPerOp)
	fmt.Printf("total throughput: %.2f ops/sec\n", totalThroughput)
}
