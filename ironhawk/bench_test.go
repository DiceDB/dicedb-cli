package ironhawk

import (
	"fmt"
	"testing"

	"github.com/DiceDB/dicedb-cli/wire"
)

func BenchmarkCommand(b *testing.B) {
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

// Benchmark with multiple connections in parallel
func BenchmarkParallelCommands(b *testing.B) {
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		conn := NewConn("localhost", 7379)
		if conn == nil {
			b.Fatal("Failed to create connection")
		}
		defer conn.Close()

		cmd := &wire.Command{
			Cmd:  "GET", // Example command - adjust as needed
			Args: []string{"key"},
		}

		for pb.Next() {
			if err := Write(conn, cmd); err != nil {
				b.Fatalf("Error sending command: %v", err)
			}
			if _, err := Read(conn); err != nil {
				b.Fatalf("Error reading response: %v", err)
			}
		}
	})
}
