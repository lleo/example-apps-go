package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/lleo/example-apps-go/netapp-grpc-protobuf/keyval"
	"google.golang.org/grpc"
)

var _ = fmt.Printf
var _ = log.Printf

const (
	defaultAddr = "localhost:9090"
)

func main() {
	addr := flag.String("a", defaultAddr, "<host>:<port>")
	flag.Parse()

	nargs := flag.NArg()
	fmt.Printf("nargs=%d\n", nargs)
	if nargs < 1 {
		usage(os.Stderr, 1, "no command given.")
	}

	cmd := flag.Arg(0)
	fmt.Printf("cmd=%s\n", cmd)
	var key string
	var val int32
	switch cmd {
	case "keys":
		if nargs != 1 {
			usage(os.Stderr, 1, "to many arguments.")
		}
	case "load":
		if nargs != 2 {
			usage(os.Stderr, 1, "incorrect number of args: expected load <key>")
		}
		key = flag.Arg(1)
	case "store":
		if nargs != 3 {
			usage(os.Stderr, 1, "incorrect number of args: expected load <key> <val>")
		}
		key = flag.Arg(1)
		vala := flag.Arg(2)
		val64, err := strconv.ParseInt(vala, 10, 32)
		if err != nil {
			usage(os.Stderr, 1, err.Error())
		}
		val = int32(val64)
	default:
		usage(os.Stderr, 1, "unknown command.")
	}

	sk, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer sk.Close()

	client := keyval.NewKeyValSvcClient(sk)

	ctx, cancelTO := context.WithTimeout(context.Background(), time.Second)
	defer cancelTO()

	switch cmd {
	case "keys":
		rsp, err := client.Keys(ctx, &keyval.KeysReq{})
		if err != nil {
			log.Fatalf("grpc Keys() failed: %v", err)
		}

		log.Printf("rsp.Keys = %v", rsp.Keys)
	case "load":
		rsp, err := client.Load(ctx, &keyval.LoadReq{Key: key})
		if err != nil {
			log.Fatalf("grpc Load(%q) failed: %v", key, err)
		}
		fmt.Printf("grpc Load(%q) -> val: %d, exists: %t\n",
			key, rsp.Val, rsp.Exists)
	case "store":
		rsp, err := client.Store(ctx, &keyval.StoreReq{Key: key, Val: val})
		if err != nil {
			log.Fatalf("grpc Store(%q, %d) failed: %v", key, val, err)
		}
		fmt.Printf("grps Store(%q, %d) -> added: %t\n", key, val, rsp.Added)
	default:
		fmt.Printf("WTF!")
	}
}

func usage(out io.Writer, xit int, msg string) {
	fmt.Fprintln(out, msg)
	fmt.Fprintln(out, "Usage: cli [option] <cmd> [args...]")
	fmt.Fprintln(out, " cmd := keys | load | store")
	fmt.Fprintln(out, " 'keys' requires no arguments.")
	fmt.Fprintln(out, " 'load' requires a key argument.")
	fmt.Fprintln(out, " 'store' requires a key and a value argument.")
	os.Exit(xit)
}
