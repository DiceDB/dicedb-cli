package ironhawk

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/chzyer/readline"
	"github.com/dicedb/dicedb-go"
	"github.com/dicedb/dicedb-go/wire"
	"github.com/fatih/color"
)

type Mode string

const (
    CommandMode Mode = "command"
    WatchMode   Mode = "watch"
)

var (
    boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
    boldRed   = color.New(color.FgRed, color.Bold).SprintFunc()
    boldBlue  = color.New(color.FgBlue, color.Bold).SprintFunc()
    currentMode = CommandMode
)

func Run(host string, port int) {
	client, err := dicedb.NewClient(host, port)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:      fmt.Sprintf("%s:%s> ", boldBlue(host), boldBlue(port)),
		HistoryFile: os.ExpandEnv("$HOME/.dicedb_history"),
	})
	if err != nil {
		fmt.Printf("%s failed to initialize readline: %v\n", boldRed("ERR"), err)
		return
	}
	defer rl.Close()

	// Setup signal handling for main CLI
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Handle Ctrl+C in a separate goroutine for the main CLI
	go func() {
		for range sigChan {
			if currentMode == CommandMode {
				fmt.Println("Received interrupt. Exiting...")
				os.Exit(0)
			}
		}
	}()

	for {
		input, err := rl.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			break
		}
		input = strings.TrimSpace(input)

		if input == "exit" {
			return
		}

		if input == "" {
			continue
		}

		args := parseArgs(input)
		if len(args) == 0 {
			continue
		}

		c := &wire.Command{
			Cmd:  strings.ToUpper(args[0]),
			Args: args[1:],
		}

		resp := client.Fire(c)

		if strings.HasSuffix(strings.ToUpper(args[0]), ".WATCH") {
			fmt.Println("Entered watch mode for", c.Cmd, strings.Join(c.Args, " "))

			// Create a watch-specific signal channel
			watchSigChan := make(chan os.Signal, 1)
			signal.Notify(watchSigChan, os.Interrupt)

			ch, err := client.WatchCh()
			if err != nil {
				fmt.Println("Error watching:", err)
				continue
			}

			// resp, _ := <-ch

			var fingerprint string
			if len(resp.Attrs.AsMap()) > 0 {
				for k, v := range resp.Attrs.AsMap() {
					if k == "fingerprint" {
						fingerprint = v.(string) // Extract fingerprint
					}
				}
			}

			// Set watch mode flag
			currentMode = WatchMode

			// Goroutine to listen to watch channel
			go func() {
				for {
					select {
					case resp, ok := <-ch:
						if !ok {
							fmt.Println("Watch channel closed")
							return
						}
						renderResponse(resp)
					}
				}
			}()

			// Wait until watch mode exits
			<-watchSigChan
			fmt.Println("Exiting watch mode...")
    		currentMode = CommandMode
    		signal.Stop(watchSigChan)

			unwatchc := &wire.Command{
				Cmd:  "UNWATCH",
				Args: []string{fingerprint},
			}

			client.Fire(unwatchc)

		} else {
			renderResponse(resp)
		}
	}
}

func renderResponse(resp *wire.Response) {
	if resp.Err != "" {
		fmt.Printf("%s %s\n", boldRed("ERR"), resp.Err)
		return
	}

	fmt.Printf("%s ", boldGreen("OK"))
	if len(resp.Attrs.AsMap()) > 0 {
		attrs := []string{}
		for k, v := range resp.Attrs.AsMap() {
			attrs = append(attrs, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Printf("[%s] ", strings.Join(attrs, ", "))
	}

	switch resp.Value.(type) {
	case *wire.Response_VStr:
		fmt.Printf("%s\n", resp.Value.(*wire.Response_VStr).VStr)
	case *wire.Response_VInt:
		fmt.Printf("%d\n", resp.Value.(*wire.Response_VInt).VInt)
	case *wire.Response_VFloat:
		fmt.Printf("%f\n", resp.Value.(*wire.Response_VFloat).VFloat)
	case *wire.Response_VBytes:
		fmt.Printf("%s\n", resp.Value.(*wire.Response_VBytes).VBytes)
	case *wire.Response_VNil:
		fmt.Printf("(nil)\n")
	}
}

func parseArgs(input string) []string {
	var args []string
	var currentArg string
	inQuotes := false
	var quoteChar byte = '"'

	for i := 0; i < len(input); i++ {
		c := input[i]
		if c == ' ' && !inQuotes {
			if currentArg != "" {
				args = append(args, currentArg)
				currentArg = ""
			}
		} else if (c == '"' || c == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = c
		} else if c == quoteChar && inQuotes {
			inQuotes = false
		} else {
			currentArg += string(c)
		}
	}
	if currentArg != "" {
		args = append(args, currentArg)
	}
	return args
}
