package ironhawk

import (
	"fmt"
	"net"
	"time"

	"github.com/DiceDB/dicedb-cli/wire"
	"google.golang.org/protobuf/proto"
)

func Run(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", addr, err)
	}
	defer conn.Close()
	fmt.Println("connected to ", addr)

	cmd := &wire.Command{
		Cmd:  "example",
		Args: []string{"arg1", "arg2"},
	}

	// Serialize the command
	data, err := proto.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %v", err)
	}

	// Send the command
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send command: %v", err)
	}

	// Read response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Unmarshal response
	resp := &wire.Response{}
	err = proto.Unmarshal(buf[:n], resp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Handle response
	if !resp.Success {
		return fmt.Errorf("command failed: %s (error: %s)", resp.Msg, resp.Err)
	}

	fmt.Printf("Command succeeded: %s\n", resp.Msg)
	return nil
}
