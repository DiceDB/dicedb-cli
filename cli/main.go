package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/dicedb/dicedb-go"
)

type DiceDBClient struct {
	client     *dicedb.Client
	subscribed bool
	subType    string
	watchConn  *dicedb.WatchConn
	subCtx     context.Context
	subCancel  context.CancelFunc
	addr       string
	password   string
	wg         sync.WaitGroup // Add WaitGroup for synchronization
}

func Run(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	password := ""
	ctx := context.Background()

	// Create a dicedb client
	client := dicedb.NewClient(&dicedb.Options{
		Addr:     addr,
		Password: password,
	})

	// Ping to test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to DiceDB: %v", err)
	}

	// Create a DiceDBClient instance
	dicedbClient := &DiceDBClient{
		client:   client,
		addr:     addr,
		password: password,
	}

	// Start the prompt
	p := prompt.New(
		dicedbClient.Executor,
		dicedbClient.Completer,
		prompt.OptionPrefix(fmt.Sprintf("dicedb (%s)> ", addr)),
		prompt.OptionLivePrefix(dicedbClient.LivePrefix),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				if dicedbClient.subscribed {
					dicedbClient.handleWatchModeExit()
				} else {
					handleExit()
				}
			},
		}),
	)
	p.Run()
	handleExit()
}

func (c *DiceDBClient) LivePrefix() (string, bool) {
	if c.subscribed {
		if c.subType != "" {
			return "", true
		}
		return fmt.Sprintf("dicedb (%s) [subscribed]> ", c.addr), true
	}
	return fmt.Sprintf("dicedb (%s)> ", c.addr), false
}

func (c *DiceDBClient) Executor(in string) {
	ctx := context.Background()

	// Do not execute anything if watch mode is on.
	if c.subscribed {
		return
	}

	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	if in == "exit" {
		handleExit()
	}

	args := parseArgs(in)
	if len(args) == 0 {
		return
	}

	cmd := strings.ToUpper(args[0])

	switch {
	case cmd == CmdAuth:
		c.handleAuth(args, ctx)

	case cmd == CmdSubscribe:
		c.handleSubscribe(args)

	case cmd == CmdUnsubscribe:
		c.handleUnsubscribe()

	default:
		// Handle custom .WATCH commands
		if strings.HasSuffix(cmd, SuffixWatch) {
			c.handleWatchCommand(cmd, args)
			return
		}

		// Handle custom .UNWATCH commands
		if strings.HasSuffix(cmd, SuffixUnwatch) {
			c.handleUnwatchCommand(args, ctx, cmd)
			return
		}

		// Execute other commands
		res, err := c.client.Do(ctx, toArgInterface(args)...).Result()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		c.printReply(res)
	}
}

func toArgInterface(args []string) []interface{} {
	argsInterface := make([]interface{}, len(args))
	for i, v := range args {
		argsInterface[i] = v
	}
	return argsInterface
}

func (c *DiceDBClient) handleUnwatchCommand(args []string, ctx context.Context, cmd string) {
	// TODO: Add error handling when the SDK does not throw an error on every unsubscribe
	err := c.watchConn.Unwatch(ctx, strings.TrimSuffix(cmd, SuffixUnwatch), args[1])
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	c.printReply("OK")
	c.subCancel()
	c.subscribed = false
	c.subType = ""
}

// TODO: Ideally this should only unwatch if the supplied fingerprint is correct.
func (c *DiceDBClient) handleWatchCommand(cmd string, args []string) {
	if c.subscribed {
		fmt.Println("Cannot execute commands while in subscription mode. Use the corresponding unsubscribe command to exit.")
		return
	}
	c.subscribed = true
	c.subType = cmd
	c.subCtx, c.subCancel = context.WithCancel(context.Background())

	baseCmd := strings.TrimSuffix(cmd, SuffixWatch)

	go c.watchCommand(baseCmd, toArgInterface(args[1:])...)
}

func (c *DiceDBClient) handleWatchModeExit() {
	fmt.Println("Exiting watch mode.")

	c.subCancel()
	c.wg.Wait()
}

func (c *DiceDBClient) handleUnsubscribe() {
	if !c.subscribed || c.subType != CmdSubscribe {
		fmt.Println("Not subscribed to any channels.")
		return
	}
	c.subCancel()
	c.subscribed = false
	c.subType = ""
}

func (c *DiceDBClient) handleSubscribe(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: SUBSCRIBE channel [channel ...]")
		return
	}
	if c.subscribed {
		fmt.Println("Already in a subscribed or watch state. Unsubscribe first.")
		return
	}
	c.subscribed = true
	c.subType = CmdSubscribe
	c.subCtx, c.subCancel = context.WithCancel(context.Background())
	go c.subscribe(args[1:])
}

func (c *DiceDBClient) handleAuth(args []string, ctx context.Context) {
	if len(args) != 2 {
		fmt.Println("Usage: AUTH password")
		return
	}
	c.password = args[1]
	// Reconnect with new password
	c.client = dicedb.NewClient(&dicedb.Options{
		Addr:     c.addr,
		Password: c.password,
	})
	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("AUTH failed: %v\n", err)
		return
	}
	fmt.Println("OK")
}

func (c *DiceDBClient) Completer(d prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{}

	if c.subscribed {
		return suggestions
	}
	// Get the text before the cursor
	beforeCursor := d.TextBeforeCursor()
	words := strings.Fields(beforeCursor)

	// Only suggest commands if we're at the first word
	if len(words) > 1 {
		return nil
	}

	text := d.GetWordBeforeCursor()
	if len(text) == 0 {
		return nil
	}

	for _, cmd := range dicedbCommands {
		if strings.HasPrefix(strings.ToUpper(cmd), strings.ToUpper(text)) {
			suggestions = append(suggestions, prompt.Suggest{Text: cmd})
		}
	}
	return suggestions
}

func (c *DiceDBClient) printReply(reply interface{}) {
	const grey = "\033[2m"
	const reset = "\033[0m"

	switch v := reply.(type) {
	case string:
		fmt.Printf("%s(string)%s \"%s\"\n", grey, reset, v)
	case int64:
		fmt.Printf("%s(integer)%s %d\n", grey, reset, v)
	case float64:
		fmt.Printf("%s(float)%s %f\n", grey, reset, v)
	case []byte:
		fmt.Printf("%s(string)%s %s\n", grey, reset, string(v))
	case []interface{}:
		fmt.Printf("%s(array):%s\n", grey, reset)
		for i, e := range v {
			fmt.Printf("  %d) ", i+1)
			c.printReply(e)
		}
	case nil:
		fmt.Printf("%s(nil)%s\n", grey, reset)
	case error:
		fmt.Printf("%s(error)%s %v\n", grey, reset, v)
	default:
		fmt.Printf("%s(unknown type)%s %v\n", grey, reset, v)
	}
}

func (c *DiceDBClient) printWatchResult(res *dicedb.WatchResult) {
	c.printReply([]interface{}{
		"fingerprint",
		res.Fingerprint,
		"data",
		res.Data,
	})
}

func (c *DiceDBClient) subscribe(channels []string) {
	defer func() {
		c.subscribed = false
		c.subType = ""
	}()

	pubsub := c.client.Subscribe(c.subCtx, channels...)
	defer pubsub.Close()

	for {
		select {
		case <-c.subCtx.Done():
			return
		default:
			msg, err := pubsub.ReceiveMessage(c.subCtx)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
		}
	}
}

func (c *DiceDBClient) watchCommand(cmd string, args ...interface{}) {
	c.wg.Add(1)
	defer func() {
		c.subscribed = false
		c.subType = ""
		if c.watchConn != nil {
			c.watchConn.Close()
		}
		c.wg.Done()
	}()

	c.watchConn = c.client.WatchConn(c.subCtx)

	// Send the WATCH command
	firstMsg, err := c.watchConn.Watch(c.subCtx, cmd, args...)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Press Ctrl+C to exit watch mode.")
	cmdFingerPrint := firstMsg.Fingerprint
	c.printWatchResult(firstMsg)

	channel := c.watchConn.Channel()
	for {
		select {
		case <-c.subCtx.Done():
			err = c.watchConn.Unwatch(c.subCtx, cmd, cmdFingerPrint)
			if err != nil {
				fmt.Printf("error in unwatch: %v\n", err)
				return
			}
			return
		case res := <-channel:
			if res == nil {
				return
			}
			c.printWatchResult(res)
		}
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

func handleExit() {
	rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
	rawModeOff.Stdin = os.Stdin
	_ = rawModeOff.Run()
	os.Exit(0)
}
