package main

import (
	"fmt"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

// Function to handle user input and execute commands
func executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Split input into command and arguments
	args := strings.Split(input, " ")
	command := args[0]

	switch command {
	case "SET":
		fmt.Println("Command: SET (Set a key-value pair)")
	case "GET":
		fmt.Println("Command: GET (Get a value by key)")
	case "EXIT":
		fmt.Println("Exiting...")
		return
	default:
		fmt.Println("Unknown command:", command)
	}
}

// Function to provide suggestions for autocompletion
func completer(d prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{
		{Text: "SET", Description: "Set a key-value pair"},
		{Text: "GET", Description: "Get a value by key"},
		{Text: "EXIT", Description: "Exit the program"},
	}
	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

func main() {
	fmt.Println("Starting Redis-like CLI...")
	prompt.New(
		executor,
		completer,
	).Run()
}
