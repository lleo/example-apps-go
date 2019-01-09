// Example key-value lookup/storage client.
// example runs:
// $ go run main.go -a "localhost:9090" put foo '{"json": true}'
// $ go run main.go get foo
// $ go run main.go keys fee fie foe fum //returns key/val pairs for key that exist
// $ go run main.go keys #gets all key/val pairs
package main

import (
	"fmt"
	"os"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/lleo/example-apps-go/key-val-network-app/keyval"
	"google.golang.org/grpc"
)

const defaultPort = "9090"
const defaultIp = "localhost"

func main() {
	fmt.Println("len(os.Args) =", len(os.Args))
	fmt.Println("os.Args =", os.Args)

	var addr string
	flags.StringVar(&addr, "a", defaultIp+":"+defaultPort,
		"(hostname|ip):port (eg \"localhost:9090\")")
	addr = defaultIp + ":" + defaultPort
	if len(os.Args) > 2 {
		addr = os.Args[1] + ":" + os.Args[2]
	}

	var conn grpc.ClientConn
	var err error
	conn, err = grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithCodec(flatbuffers.FlatbuffersCodec{}),
	)
	if err != nil {
		fmt.Panicf("grpc.Dial failed: %s\n", err)
	}
	defer conn.Close()

	client := keyval.NewKeyValSvcClient(conn)

	switch cmd {
	case "put":
	case "get":
	case "keys":
		fmt.Panicf("%q cmd not implemented\n", cmd)
	default:
		fmt.Panicf("unknown cmd: %q\n", cmd)
	}
	fmt.Println("done.")
}
