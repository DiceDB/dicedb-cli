package ironhawk

import (
	"fmt"
	"testing"

	"github.com/DiceDB/dicedb-cli/wire"
)

func benchmarkCommand(b *testing.B) {
	conn := NewConn("localhost", 7379)
	if conn == nil {
		b.Fatal("Failed to create connection")
	}
	defer conn.Close()

	cmds := make([]*wire.Command, 1000)
	for i := 0; i < 1000; i++ {
		cmds[i] = &wire.Command{
			Cmd:  "GET",
			Args: []string{fmt.Sprintf("key-%d", i)},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Write(conn, cmds[i%1000]); err != nil {
			b.Fatalf("Error sending command: %v", err)
		}
		if _, err := Read(conn); err != nil {
			b.Fatalf("Error reading response: %v", err)
		}
	}
}

func Benchmark(parallelism int) {
	nsPerOpChan := make(chan float64, parallelism)
	allocsPerOpChan := make(chan int64, parallelism)
	bytesPerOpChan := make(chan int64, parallelism)
	throughputChan := make(chan float64, parallelism)

	for i := 0; i < parallelism; i++ {
		go func() {
			results := testing.Benchmark(benchmarkCommand)
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
