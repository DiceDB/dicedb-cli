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
	"github.com/google/shlex"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	boldRed   = color.New(color.FgRed, color.Bold).SprintFunc()
	boldBlue  = color.New(color.FgBlue, color.Bold).SprintFunc()
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
		if resp.Err != "" {
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

	if len(resp.VSsMap) > 0 {
		fmt.Println()
		for k, v := range resp.VSsMap {
			fmt.Printf("%s=%s\n", k, v)
		}
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

	if len(resp.GetVList()) > 0 {
		fmt.Println()
		for i, v := range resp.GetVList() {
			//TODO: handle structpb.Value_StructValue & structpb.Value_ListValue
			switch v.GetKind().(type) {
			case *structpb.Value_NullValue:
				fmt.Printf("%d) (nil)\n", i+1)
			case *structpb.Value_NumberValue:
				fmt.Printf("%d) %f\n", i+1, v.GetNumberValue())
			case *structpb.Value_StringValue:
				fmt.Printf("%d) \"%s\"\n", i+1, v.GetStringValue())
			case *structpb.Value_BoolValue:
				fmt.Printf("%d) %t\n", i+1, v.GetBoolValue())
			}
		}
	}
}

func parseArgs(input string) []string {
	args, err := shlex.Split(input)
	if err != nil {
		fmt.Printf("%s failed to parse command: %v\n", boldRed("ERR"), err)
		return []string{}
	}
	return args
}
