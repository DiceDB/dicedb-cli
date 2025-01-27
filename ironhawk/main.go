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

		if resp.Success {
			fmt.Printf("%s %s\n", boldGreen("OK"), resp.Msg)
		} else {
			fmt.Printf("%s %s\n", boldRed("ERR"), resp.Err)
		}
	}
}
