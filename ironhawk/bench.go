package ironhawk

import (
	"fmt"
	"testing"

	"github.com/dicedb/dicedb-cli/bench"
	"github.com/dicedb/dicedb-go"
	"github.com/dicedb/dicedb-go/wire"
)

func benchmarkCommand(b *testing.B) {
	client, err := dicedb.NewClient("localhost", 7379)
	if err != nil {
		b.Fatal("Failed to create connection")
	}
	defer client.Close()

	cmds := make([]*wire.Command, 1000)
	for i := 0; i < 1000; i++ {
		cmds[i] = &wire.Command{
			Cmd:  "GET",
			Args: []string{fmt.Sprintf("key-%d", i)},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Fire(cmds[i%1000])
	}
}

func Benchmark(parallelism int) {
	bench.Benchmark(parallelism, benchmarkCommand)
}
