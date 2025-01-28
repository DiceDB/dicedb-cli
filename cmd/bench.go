package cmd

import (
	"fmt"

	"github.com/DiceDB/dicedb-cli/ironhawk"
	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "quickly benchmark the performance of DiceDB",
	Run: func(cmd *cobra.Command, args []string) {
		numConns, _ := cmd.Flags().GetInt("num-connections")
		engine, _ := cmd.Flags().GetString("engine")
		if engine == "ironhawk" {
			ironhawk.Benchmark(numConns)
		} else {
			fmt.Println("Invalid engine")
		}
	},
}

func init() {
	benchCmd.Flags().Int("num-connections", 4, "number of connections in parallel to fire the requests")
	benchCmd.Flags().String("engine", "ironhawk", "engine to use for the benchmark: ironhawk, resp")
	rootCmd.AddCommand(benchCmd)
}
