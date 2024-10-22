package main

import (
	"dicedb-cli/internal/completer"
	"dicedb-cli/internal/connect"
	"dicedb-cli/internal/executor"
	"fmt"
	"github.com/c-bata/go-prompt"
)

func main() {
	client, err := connect.Connect()
	if err != nil {
		fmt.Println("Error connecting to Redis server:", err)
		return
	}

	prompt.New(
		func(input string) { executor.Executor(input, client) },
		completer.Completer,
	).Run()
}
