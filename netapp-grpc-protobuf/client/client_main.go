package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lleo/example-apps-go/netapp-grpc-protobuf/keyval"
	"google.golang.org/grpc"
)

var _ = fmt.Printf
var _ = log.Printf

const (
	addr = "localhost:9090"
)

func main() {
	sk, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer sk.Close()

	client := keyval.NewKeyValSvcClient(sk)

	ctx, cancelTO := context.WithTimeout(context.Background(), time.Second)
	defer cancelTO()

	rsp, err := client.Keys(ctx, &keyval.KeysReq{})
	if err != nil {
		log.Fatalf("grpc Keys() failed: %v", err)
	}

	log.Printf("rsp.Keys = %v", rsp.Keys)
}
