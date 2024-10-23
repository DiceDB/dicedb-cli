package main

import (
	"context"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/dicedb/dicedb-go"
	"log"
	"os"
	"strings"
)

type DiceDBClient struct {
	client     *dicedb.Client
	pubsub     *dicedb.PubSub
	subscribed bool
	channels   []string
	addr       string
	password   string
}

func main() {
	addr := "localhost:7379"
	password := "" // Set this from command line arguments or environment variables if needed
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
	rclient := &DiceDBClient{
		client:   client,
		addr:     addr,
		password: password,
	}

	// Start the prompt
	fmt.Println("Connected to DiceDB. Type 'exit' or press Ctrl+D to exit.")
	p := prompt.New(
		rclient.Executor,
		rclient.Completer,
		prompt.OptionPrefix("dicedb> "),
		prompt.OptionLivePrefix(rclient.LivePrefix),
	)
	p.Run()
}

func (c *DiceDBClient) LivePrefix() (string, bool) {
	if c.subscribed {
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

	if c.subscribed && !strings.HasPrefix(strings.ToUpper(in), "UNSUBSCRIBE") {
		fmt.Println("Cannot execute commands while subscribed. Use UNSUBSCRIBE to exit subscription mode.")
		return
	}

	args := parseArgs(in)
	if len(args) == 0 {
		return
	}

	cmd := strings.ToUpper(args[0])

	switch cmd {
	case "AUTH":
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
	case "SUBSCRIBE":
		if len(args) < 2 {
			fmt.Println("Usage: SUBSCRIBE channel [channel ...]")
			return
		}
		if c.subscribed {
			fmt.Println("Already subscribed. Unsubscribe first.")
			return
		}
		c.subscribed = true
		c.channels = args[1:]
		c.pubsub = c.client.Subscribe(ctx, c.channels...)
		go c.subscribe()
	case "UNSUBSCRIBE":
		if !c.subscribed {
			fmt.Println("Not subscribed to any channels.")
			return
		}
		if len(args) > 1 {
			c.pubsub.Unsubscribe(ctx, args[1:]...)
		} else {
			c.pubsub.Unsubscribe(ctx)
		}
		c.subscribed = false
	default:
		// Convert []string to []interface{}
		argsInterface := make([]interface{}, len(args))
		for i, v := range args {
			argsInterface[i] = v
		}

		// Execute other commands
		res, err := c.client.Do(ctx, argsInterface...).Result()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		c.printReply(res)
	}
}

func (c *DiceDBClient) Completer(d prompt.Document) []prompt.Suggest {
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
	switch v := reply.(type) {
	case string:
		fmt.Println(v)
	case int64:
		fmt.Println(v)
	case []byte:
		fmt.Println(string(v))
	case []interface{}:
		for i, e := range v {
			fmt.Printf("%d) ", i+1)
			c.printReply(e)
		}
	case nil:
		fmt.Println("(nil)")
	case error:
		fmt.Printf("(error) %v\n", v)
	default:
		fmt.Printf("%v\n", v)
	}
}

func (c *DiceDBClient) subscribe() {
	ctx := context.Background()
	defer func() {
		c.subscribed = false
		c.channels = nil
		c.pubsub.Close()
	}()

	for {
		msg, err := c.pubsub.ReceiveMessage(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
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

// TODO: Ensure this list stays in sync with the actual commands supported by DiceDB.
var dicedbCommands = []string{
	"GET", "SET", "DEL", "INCR", "DECR", "EXISTS", "AUTH", "SUBSCRIBE", "UNSUBSCRIBE", "PUBLISH",
	"LPUSH", "RPUSH", "LPOP", "RPOP", "SADD", "SREM", "SMEMBERS", "HSET", "HGET", "HDEL",
	"KEYS", "PING", "QUIT", "EXPIRE", "TTL", "INFO",
}
