package cmd

import (
	"github.com/dicedb/dicedb-cli/ironhawk"
	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "quickly benchmark the performance of DiceDB",
	Run: func(cmd *cobra.Command, args []string) {
		numConns, _ := cmd.Flags().GetInt("num-connections")
		ironhawk.Benchmark(numConns)
	},
}

func init() {
	benchCmd.Flags().Int("num-connections", 4, "number of connections in parallel to fire the requests")
	rootCmd.AddCommand(benchCmd)
}
