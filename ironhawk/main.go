package ironhawk

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/DiceDB/dicedb-cli/wire"
	"github.com/fatih/color"
)

var (
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
	boldRed   = color.New(color.FgRed, color.Bold).SprintFunc()
	boldBlue  = color.New(color.FgBlue, color.Bold).SprintFunc()
)

func Run(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer conn.Close()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Handle Ctrl+C in a separate goroutine
	go func() {
		<-sigChan
		fmt.Println("received interrupt. exiting...")
		os.Exit(0)
	}()

	for {
		fmt.Printf("%s:%s> ", boldBlue(host), boldBlue(port))
		var input string

		_, _ = fmt.Scanln(&input)
		input = strings.TrimSpace(input)

		if input == "exit" {
			return
		}

		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		if err := Write(conn, &wire.Command{
			Cmd:  strings.ToUpper(args[0]),
			Args: args[1:],
		}); err != nil {
			fmt.Printf("%s failed to send command: %v\n", boldRed("ERR"), err)
			continue
		}

		resp, err := Read(conn)
		if err != nil {
			fmt.Printf("%s failed to read response: %v\n", boldRed("ERR"), err)
			continue
		}

		if resp.Err == "" {
			switch resp.Value.(type) {
			case *wire.Response_VStr:
				fmt.Printf("%s %s\n", boldGreen("OK"), resp.Value.(*wire.Response_VStr).VStr)
			case *wire.Response_VInt:
				fmt.Printf("%s %d\n", boldGreen("OK"), resp.Value.(*wire.Response_VInt).VInt)
			case *wire.Response_VFloat:
				fmt.Printf("%s %f\n", boldGreen("OK"), resp.Value.(*wire.Response_VFloat).VFloat)
			case *wire.Response_VBytes:
				fmt.Printf("%s %v\n", boldGreen("OK"), resp.Value.(*wire.Response_VBytes).VBytes)
			case *wire.Response_VNil:
				fmt.Printf("%s %s\n", boldGreen("OK"), "(nil)")
			}
		} else {
			fmt.Printf("%s %s\n", boldRed("ERR"), resp.Err)
		}
	}
}
