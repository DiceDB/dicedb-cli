package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
	fmt.Println("Connected to DiceDB. Type 'exit' or press Ctrl+D to exit.")
	p := prompt.New(
		dicedbClient.Executor,
		dicedbClient.Completer,
		prompt.OptionPrefix("dicedb> "),
		prompt.OptionLivePrefix(dicedbClient.LivePrefix),
	)
	p.Run()
}

func (c *DiceDBClient) LivePrefix() (string, bool) {
	if c.subscribed {
		if c.subType != "" {
			return fmt.Sprintf("dicedb(%s)> ", strings.ToLower(c.subType)), true
		}
		return "dicedb(subscribed)> ", true
	}
	return "dicedb> ", false
}

func (c *DiceDBClient) Executor(in string) {
	ctx := context.Background()
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}
	if in == "exit" {
		os.Exit(0)
	}

	// Prevent executing other commands while subscribed
	if c.subscribed && !c.isAllowedDuringSubscription(in) {
		fmt.Println("Cannot execute commands while in subscription mode. Use the corresponding unsubscribe command to exit.")
		return
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
	err := c.watchConn.Unwatch(ctx, strings.TrimSuffix(cmd, SuffixUnwatch), toArgInterface(args[1:]))
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
		fmt.Println("Already in a subscribed or watch state. Unsubscribe first.")
		return
	}
	c.subscribed = true
	c.subType = cmd
	c.subCtx, c.subCancel = context.WithCancel(context.Background())

	// Extract the base command
	baseCmd := strings.TrimSuffix(cmd, SuffixWatch)

	go c.watchCommand(baseCmd, toArgInterface(args[1:])...)
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

	suggestions := []prompt.Suggest{}
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
		fmt.Printf("%s(string)%s %s\n", grey, reset, v)
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
	fmt.Printf("Command: %s\n", res.Command)
	fmt.Printf("Fingerprint: %s\n", res.Fingerprint)
	fmt.Printf("Data: %v\n", res.Data)
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
	defer func() {
		c.subscribed = false
		c.subType = ""
	}()

	c.watchConn = c.client.WatchConn(c.subCtx)
	defer c.watchConn.Close()

	// Send the WATCH command
	firstMsg, err := c.watchConn.Watch(c.subCtx, cmd, args...)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the first message
	c.printWatchResult(firstMsg)

	channel := c.watchConn.Channel()

	for {
		select {
		case <-c.subCtx.Done():
			return
		case res := <-channel:
			if res == nil {
				return
			}
			c.printWatchResult(res)
		}
	}
}

func (c *DiceDBClient) isAllowedDuringSubscription(input string) bool {
	args := parseArgs(input)
	if len(args) == 0 {
		return false
	}

	cmd := strings.ToUpper(args[0])

	// Allow UNSUBSCRIBE or the corresponding .UNWATCH command during subscription
	if cmd == CmdUnsubscribe && c.subType == CmdSubscribe {
		return true
	}
	if strings.HasSuffix(c.subType, SuffixWatch) && cmd == strings.Replace(c.subType, SuffixWatch, SuffixUnwatch, 1) {
		return true
	}
	return false
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
