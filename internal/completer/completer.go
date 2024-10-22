package completer

import "github.com/c-bata/go-prompt"

// Function to provide suggestions for autocompletion
func Completer(d prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{
		{Text: "SET", Description: "Set a key-value pair"},
		{Text: "GET", Description: "Get a value by key"},
		{Text: "EXIT", Description: "Exit the program"},
	}
	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}
