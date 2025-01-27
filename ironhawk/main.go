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
	cmd := &wire.Command{
		Cmd:  "PING",
		Args: []string{},
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

	resp, err := Read(conn)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Println(resp.Msg)
	return nil
}
