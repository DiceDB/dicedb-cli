package connect

import (
	"dicedb-cli/internal/env"
	"flag"
	"fmt"
	"github.com/holys/goredis"
)

var (
	hostname = flag.String("h", env.GetEnv("REDIS_HOST", "127.0.0.1"), "Server hostname")
	port     = flag.String("p", env.GetEnv("REDIS_PORT", "7379"), "Server server port")
	socket   = flag.String("s", "", "Server socket. (overwrites hostname and port)")
	auth     = flag.String("a", "", "Password to use when connecting to the server")
)

var client *goredis.Client

func Connect() (*goredis.Client, error) {
	var err error = nil

	if client == nil {
		addr := addr()

		fmt.Println("Connecting...", addr)

		client = goredis.NewClient(addr, "")
		client.SetMaxIdleConns(1)

		err = sendPing(client)
		if err != nil {
			fmt.Println("Failed to send ping to server", err)
			return nil, err
		}

		err = sendAuth(client, *auth)
		if err != nil {
			fmt.Println("Failed to authenticate to server", err)
			return nil, err
		}
	}

	return client, nil
}

func addr() string {
	var addr string
	if len(*socket) > 0 {
		addr = *socket
	} else {
		addr = fmt.Sprintf("%s:%s", *hostname, *port)
	}
	return addr
}

func sendAuth(client *goredis.Client, passwd string) error {
	if passwd == "" {
		// do nothing
		return nil
	}

	resp, err := client.Do("AUTH", passwd)
	if err != nil {
		fmt.Printf("(error) %s\n", err.Error())
		return err
	}

	switch resp := resp.(type) {
	case goredis.Error:
		fmt.Printf("(error) %s\n", resp.Error())
		return resp
	}

	return nil
}

func sendPing(client *goredis.Client) error {
	_, err := client.Do("PING")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return err
	}
	return nil
}
