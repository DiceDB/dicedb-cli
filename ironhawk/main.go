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

type CLI struct {
	client      *dicedb.Client
	rl          *readline.Instance
	mode        Mode
	boldGreen   func(a ...interface{}) string
	boldRed     func(a ...interface{}) string
	boldBlue    func(a ...interface{}) string
	sigChan     chan os.Signal
	watchSigChan chan os.Signal
}

func NewCLI(host string, port int) (*CLI, error) {
	client, err := dicedb.NewClient(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:      fmt.Sprintf("%s:%s> ", color.BlueString(host), color.BlueString(fmt.Sprint(port))),
		HistoryFile: os.ExpandEnv("$HOME/.dicedb_history"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize readline: %v", err)
	}

	return &CLI{
		client:      client,
		rl:          rl,
		mode:        CommandMode,
		boldGreen:   color.New(color.FgGreen, color.Bold).SprintFunc(),
		boldRed:     color.New(color.FgRed, color.Bold).SprintFunc(),
		boldBlue:    color.New(color.FgBlue, color.Bold).SprintFunc(),
		sigChan:     make(chan os.Signal, 1),
		watchSigChan: make(chan os.Signal, 1),
	}, nil
}

func Run(host string, port int) {
	cli, err := NewCLI(host, port)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.cleanup()

	cli.setupSignalHandler()
	cli.run()
}

func (c *CLI) cleanup() {
	c.client.Close()
	c.rl.Close()
}

func (c *CLI) setupSignalHandler() {
	signal.Notify(c.sigChan, os.Interrupt)
	go func() {
		for range c.sigChan {
			if c.mode == CommandMode {
				fmt.Println("\nReceived interrupt. Exiting...")
				os.Exit(0)
			}
		}
	}()
}

func (c *CLI) run() {
	for {
		input, err := c.rl.Readline()
		if err != nil {
			break
		}

		if !c.handleInput(strings.TrimSpace(input)) {
			return
		}
	}
}

func (c *CLI) handleInput(input string) bool {
	if input == "exit" {
		return false
	}

	if input == "" {
		return true
	}

	args := parseArgs(input)
	if len(args) == 0 {
		return true
	}

	cmd := &wire.Command{
		Cmd:  strings.ToUpper(args[0]),
		Args: args[1:],
	}

	resp := c.client.Fire(cmd)

	if strings.HasSuffix(strings.ToUpper(args[0]), ".WATCH") {
		c.handleWatchMode(cmd, resp)
	} else {
		c.renderResponse(resp)
	}

	return true
}

func (c *CLI) handleWatchMode(cmd *wire.Command, resp *wire.Response) {
	fmt.Printf("Entered watch mode for %s %s\n", cmd.Cmd, strings.Join(cmd.Args, " "))

	signal.Notify(c.watchSigChan, os.Interrupt)
	defer signal.Stop(c.watchSigChan)

	ch, err := c.client.WatchCh()
	if err != nil {
		fmt.Printf("Error watching: %v\n", err)
		return
	}

	fingerprint := c.extractFingerprint(resp)
	c.mode = WatchMode

	done := make(chan struct{})
	go c.watchLoop(ch, done)

	<-c.watchSigChan
	fmt.Println("\nExiting watch mode...")
	close(done)
	c.mode = CommandMode

	unwatchCmd := &wire.Command{
		Cmd:  "UNWATCH",
		Args: []string{fingerprint},
	}
	c.client.Fire(unwatchCmd)
}

func (c *CLI) watchLoop(ch <-chan *wire.Response, done chan struct{}) {
	for {
		select {
		case resp, ok := <-ch:
			if !ok {
				fmt.Println("Watch channel closed")
				return
			}
			c.renderResponse(resp)
		case <-done:
			return
		}
	}
}

func (c *CLI) extractFingerprint(resp *wire.Response) string {
	if attrs := resp.Attrs.AsMap(); len(attrs) > 0 {
		if fp, ok := attrs["fingerprint"].(string); ok {
			return fp
		}
	}
	return ""
}

func (c *CLI) renderResponse(resp *wire.Response) {
	if resp.Err != "" {
		fmt.Printf("%s %s\n", c.boldRed("ERR"), resp.Err)
		return
	}

	fmt.Printf("%s ", c.boldGreen("OK"))
	if len(resp.Attrs.AsMap()) > 0 {
		attrs := make([]string, 0, len(resp.Attrs.AsMap()))
		for k, v := range resp.Attrs.AsMap() {
			attrs = append(attrs, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Printf("[%s] ", strings.Join(attrs, ", "))
	}

	switch v := resp.Value.(type) {
	case *wire.Response_VStr:
		fmt.Printf("%s\n", v.VStr)
	case *wire.Response_VInt:
		fmt.Printf("%d\n", v.VInt)
	case *wire.Response_VFloat:
		fmt.Printf("%f\n", v.VFloat)
	case *wire.Response_VBytes:
		fmt.Printf("%s\n", v.VBytes)
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
