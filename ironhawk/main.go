package ironhawk

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/DiceDB/dicedb-cli/wire"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

var (
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	boldRed   = color.New(color.FgRed, color.Bold).SprintFunc()
	boldBlue  = color.New(color.FgBlue, color.Bold).SprintFunc()
)

func NewConn(host string, port int) net.Conn {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	return conn
}

func Run(host string, port int) {
	conn := NewConn(host, port)
	if conn == nil {
		fmt.Println("Failed to create connection")
		return
	}
	defer conn.Close()

	// Initialize readline
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      fmt.Sprintf("%s:%s> ", boldBlue(host), boldBlue(port)),
		HistoryFile: os.ExpandEnv("$HOME/.dicedb_history"),
	})
	if err != nil {
		fmt.Printf("%s failed to initialize readline: %v\n", boldRed("ERR"), err)
		return
	}
	defer rl.Close()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Handle Ctrl+C in a separate goroutine
	go func() {
		<-sigChan
		fmt.Println("\nreceived interrupt. exiting...")
		os.Exit(0)
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

		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		c := &wire.Command{
			Cmd:  strings.ToUpper(args[0]),
			Args: args[1:],
		}
		if err := Write(conn, c); err != nil {
			fmt.Printf("%s failed to send command: %v\n", boldRed("ERR"), err)
			continue
		}

		resp, err := Read(conn)
		if err != nil {
			fmt.Printf("%s failed to read response: %v\n", boldRed("ERR"), err)
			continue
		}

		if resp.Err != "" {
			fmt.Printf("%s %s\n", boldRed("ERR"), resp.Err)
			continue
		}

		if strings.HasSuffix(c.Cmd, ".WATCH") {
			fmt.Println("entered the watch mode for", c.Cmd, strings.Join(c.Args, " "))
			for {
				renderResponse(resp)
				resp, err = Read(conn)
				if err != nil {
					fmt.Printf("%s failed to read response: %v\n", boldRed("ERR"), err)
					break
				}
			}
		} else {
			renderResponse(resp)
		}
	}
}

func renderResponse(resp *wire.Response) {
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
