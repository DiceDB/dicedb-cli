package silverpine

import (
	"context"
	"fmt"
	"testing"

	"github.com/DiceDB/dicedb-cli/bench"
	"github.com/dicedb/dicedb-go"
)

func NewConn(host string, port int) *dicedb.Client {
	client := dicedb.NewClient(&dicedb.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})
	return client
}

func benchmarkCommand(b *testing.B) {
	conn := NewConn("localhost", 7379)
	if conn == nil {
		b.Fatal("Failed to create connection")
	}
	defer conn.Close()

	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
	}

	_, _ = conn.Get(context.Background(), keys[0]).Result()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = conn.Get(context.Background(), keys[i%1000]).Result()
	}
}

func Benchmark(parallelism int) {
	bench.Benchmark(parallelism, benchmarkCommand)
}
