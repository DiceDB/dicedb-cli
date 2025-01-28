package ironhawk

import (
	"fmt"
	"testing"

	"github.com/DiceDB/dicedb-cli/bench"
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
	bench.Benchmark(parallelism, benchmarkCommand)
}
