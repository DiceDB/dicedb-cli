package ironhawk

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/chzyer/readline"
	"github.com/dicedb/dicedb-go"
	"github.com/dicedb/dicedb-go/wire"
	"github.com/fatih/color"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	boldRed   = color.New(color.FgRed, color.Bold).SprintFunc()
	boldBlue  = color.New(color.FgBlue, color.Bold).SprintFunc()
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
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

	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	watchModeSignal := make(chan bool, 1)
	sigChanWatchMode := make(chan os.Signal, 1)

	// Handling interrupts in a goroutine
	go func() {
		for sig := range sigChan {
			select {
			// When in watch mode, capture the signal and send it to
			// the signal channel for watch mode
			case <-watchModeSignal:
				// Instead of exiting the REPL, send the signal to the
				// watch mode signal channel
				sigChanWatchMode <- sig
			default:
				// when not in watch mode, exit the REPL
				fmt.Println("\nreceived interrupt. exiting...")
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
		if resp.Status == wire.Status_ERR {
			renderResponse(resp)
			continue
		}

		if strings.HasSuffix(strings.ToUpper(args[0]), ".WATCH") {
			fmt.Println("entered the watch mode for", c.Cmd, strings.Join(c.Args, " "))

			// Send a signal to the primary Signal handler goroutine
			// that the watch mode has been entered
			watchModeSignal <- true

			// Get the watch channel and start watching for changes
			ch, err := client.WatchCh()
			if err != nil {
				fmt.Println("error watching:", err)
				continue
			}

			// Start watching for changes
			// until the user exits the watch mode
			shouldExitWatchMode := false
			for !shouldExitWatchMode {
				select {
				// If the user sends a signal Ctrl+C,
				// It is captured by the signal handler goroutine
				// and then sent to the watch mode signal channel
				// which will set the shouldExitWatchMode flag to true
				case <-sigChanWatchMode:
					fmt.Println("exiting the watch mode. back to command mode")
					shouldExitWatchMode = true
				case resp := <-ch:
					// If we get any response over the watch channel,
					// render the response
					renderResponse(resp)
				}
			}
		} else {
			// If the command is not a watch command, render the response
			// and continue to the next command in REPL
			renderResponse(resp)
		}
	}
}

func renderResponse(resp *wire.Result) {
	if resp.Status == wire.Status_ERR {
		fmt.Printf("%s %s\n", boldRed("ERR"), resp.Message)
		return
	}

	switch resp.Response.(type) {
	case *wire.Result_GETRes:
		fmt.Printf("%s \"%s\"\n", boldGreen(resp.Message), resp.GetGETRes().Value)
	case *wire.Result_GETDELRes:
		fmt.Printf("%s \"%s\"\n", boldGreen(resp.Message), resp.GetGETDELRes().Value)
	case *wire.Result_SETRes:
		fmt.Printf("%s\n", boldGreen(resp.Message))
	case *wire.Result_FLUSHDBRes:
		fmt.Printf("%s\n", boldGreen(resp.Message))
	case *wire.Result_DELRes:
		fmt.Printf("%s %d\n", boldGreen(resp.Message), resp.GetDELRes().Count)
	case *wire.Result_DECRRes:
		fmt.Printf("%s %d\n", boldGreen(resp.Message), resp.GetDECRRes().Value)
	case *wire.Result_INCRRes:
		fmt.Printf("%s %d\n", boldGreen(resp.Message), resp.GetINCRRes().Value)
	case *wire.Result_DECRBYRes:
		fmt.Printf("%s %d\n", boldGreen(resp.Message), resp.GetDECRBYRes().Value)
	case *wire.Result_INCRBYRes:
		fmt.Printf("%s %d\n", boldGreen(resp.Message), resp.GetINCRBYRes().Value)
	case *wire.Result_ECHORes:
		fmt.Printf("%s %s\n", boldGreen(resp.Message), resp.GetECHORes().Message)
	case *wire.Result_EXISTSRes:
		fmt.Printf("%s %d\n", boldGreen(resp.Message), resp.GetEXISTSRes().Count)
	case *wire.Result_EXPIRERes:
		fmt.Printf("%s %v\n", boldGreen(resp.Message), resp.GetEXPIRERes().IsChanged)
	case *wire.Result_EXPIREATRes:
		fmt.Printf("%s %v\n", boldGreen(resp.Message), resp.GetEXPIREATRes().IsChanged)
	case *wire.Result_EXPIRETIMERes:
		fmt.Printf("%s %v\n", boldGreen(resp.Message), resp.GetEXPIRETIMERes().UnixSec)
	case *wire.Result_TTLRes:
		fmt.Printf("%s %v\n", boldGreen(resp.Message), resp.GetTTLRes().Seconds)
	case *wire.Result_GETEXRes:
		fmt.Printf("%s \"%s\"\n", boldGreen(resp.Message), resp.GetGETEXRes().Value)
	case *wire.Result_GETSETRes:
		fmt.Printf("%s \"%s\"\n", boldGreen(resp.Message), resp.GetGETSETRes().Value)
	case *wire.Result_HANDSHAKERes:
		fmt.Printf("%s\n", boldGreen(resp.Message))
	case *wire.Result_HGETRes:
		fmt.Printf("%s \"%s\"\n", boldGreen(resp.Message), resp.GetHGETRes().Value)
	case *wire.Result_HSETRes:
		fmt.Printf("%s %d\n", boldGreen(resp.Message), resp.GetHSETRes().Count)
	case *wire.Result_HGETALLRes:
		fmt.Printf("%s\n", boldGreen(resp.Message))
		for i, e := range resp.GetHGETALLRes().Elements {
			fmt.Printf("%d) %s=\"%s\"\n", i, e.Key, e.Value)
		}
	case *wire.Result_KEYSRes:
		fmt.Printf("%s\n", boldGreen(resp.Message))
		for i, key := range resp.GetKEYSRes().Keys {
			fmt.Printf("%d) %s\n", i, key)
		}
	case *wire.Result_PINGRes:
		fmt.Printf("%s \"%s\"\n", boldGreen(resp.Message), resp.GetPINGRes().Message)
	default:
		fmt.Println("note: this response is JSON serialized version of the response because it is not supported by this version of the CLI. You can upgrade the CLI to the latest version to get a formatted response.")
		b, err := protojson.Marshal(resp)
		if err != nil {
			log.Fatalf("failed to marshal to JSON: %v", err)
		}

		var m map[string]interface{}
		_ = json.Unmarshal(b, &m)

		nb, _ := json.MarshalIndent(m, "", "  ")
		fmt.Println(string(nb))
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
