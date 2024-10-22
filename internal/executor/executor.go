package executor

import (
	"dicedb-cli/internal/print"
	"fmt"
	"github.com/holys/goredis"
	"strings"
)

// Function to handle user input and execute commands
func Executor(input string, client *goredis.Client) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return
	}

	cmds := strings.Split(input, " ")
	cmd := cmds[0]

	args := make([]interface{}, len(cmds[1:]))
	for i, arg := range cmds[1:] {
		//This will remove any surrounding quotes. Example - SET "key" 'value' to ["key", "value"]
		args[i] = strings.Trim(arg, "\"'")
	}

	response, err := client.Do(cmd, args...)
	if err != nil {
		fmt.Println("Error in executing command", err)
	}

	print.PrintResponse(0, response)
}
