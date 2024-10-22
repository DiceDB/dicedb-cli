package print

import (
	"fmt"
	"github.com/holys/goredis"
	"strings"
)

func PrintResponse(level int, reply interface{}) {
	switch reply := reply.(type) {
	case int64:
		fmt.Printf("(integer) %d", reply)
	case string:
		fmt.Printf("%s", reply)
	case []byte:
		fmt.Printf("%q", reply)
	case nil:
		fmt.Printf("(nil)")
	case goredis.Error:
		fmt.Printf("(error) %s", string(reply))
	case []interface{}:
		printArrayResponse(level, reply)
	default:
		fmt.Printf("Unknown reply type: %+v", reply)
	}

	fmt.Println("\n")
}

// Helper function to print array responses
func printArrayResponse(level int, reply []interface{}) {
	for i, v := range reply {
		if i != 0 {
			fmt.Printf("%s", strings.Repeat(" ", level*4))
		}
		fmt.Printf("%-4d) ", i+1)

		PrintResponse(level+1, v)

		if i != len(reply)-1 {
			fmt.Println()
		}
	}
}
