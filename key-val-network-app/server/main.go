// Example key-value lookup/storage server.
// example runs:
// $ go run main.go -a "localhost:9090"
package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	flatbuffers "github.com/google/flatbuffers/go"
)

const defaultPort = 6060
const defaultIP = "localhost"

func main() {
	fmt.Println("len(os.Args) =", len(os.Args))
	fmt.Println("os.Args =", os.Args)
	fmt.Print("Starting...")
	svrSk, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	gSvr := grpc.NewServer(grpc.CustomCodec(
		flatbuffers.FlatbuffersCodec{}))

	//register Service
	svr := gSvr.Serve(svrSk)

	if err := svr.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
	fmt.Println("started.")
}
