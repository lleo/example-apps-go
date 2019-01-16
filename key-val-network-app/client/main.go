// Example key-value lookup/storage client.
// example runs:
// $ go run main.go -a "localhost:9090" put foo '{"json": true}'
// $ go run main.go get foo
// $ go run main.go keys fee fie foe fum //returns key/val pairs for key that exist
// $ go run main.go keys #gets all key/val pairs
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/pkg/errors"

	//"github.com/lleo/example-apps-go/key-val-network-app/keyval"
	"github.com/lleo/example-apps-go/key-val-network-app/keyval"
	"google.golang.org/grpc"
)

const defaultAddr = "localhost:9090"

func main() {
	fmt.Println("len(os.Args) =", len(os.Args))
	fmt.Println("os.Args =", os.Args)

	var addr string
	flag.StringVar(&addr, "a", defaultAddr,
		"(hostname|ip):port (eg \"localhost:9090\") or \":9090\"")
	flag.Parse()

	var args = flag.Args()
	if len(args) < 2 {
		usage(os.Stdout, 0, "")
	}

	b := flatbuffers.NewBuilder(0)

	cmd := strings.ToLower(args[0])
	var key string
	var val string
	switch cmd {
	case "get":
		if len(args) != 2 {
			usage(os.Stderr, 1, "args != 2")
		}
		key = args[1] // $ cli -a ":9090" get[0] "value"[1]
		fmt.Println("command:", cmd, key)
	case "put":
		if len(args) != 3 {
			usage(os.Stderr, 1, "args != 3")
		}
		key = args[1]
		val = args[2]
		fmt.Println("command:", cmd, key, val)
	case "getkeys":
		usage(os.Stderr, 1, "\"getkeys\" not implemented.")
	default:
		usage(os.Stderr, 1, "unknown command "+cmd)
	}

	var conn *grpc.ClientConn
	var err error
	conn, err = grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithCodec(flatbuffers.FlatbuffersCodec{}),
	)
	if err != nil {
		panic(fmt.Errorf("grpc.Dial failed: %s", err))
	}
	defer conn.Close()

	client := keyval.NewKeyValSvcClient(conn)

	switch cmd {
	case "get":
		//build request

		//must do b.CreateString() outside of keyval.*Start() and
		//keyval.*Finish() calls or else I get "Incorrect creation order:
		//object must /not be nested." error panic from keyval library.
		reqKey := b.CreateString(key)

		keyval.GetReqStart(b)
		keyval.GetReqAddKey(b, reqKey)
		b.Finish(keyval.GetReqEnd(b))

		//client call
		var rsp *keyval.GetRsp
		fmt.Printf("Sending Get(%q)\n", key)
		rsp, err = client.Get(context.Background(), b)
		if err != nil {
			panic(errors.Wrap(err, "Get grpc request failed"))
		}

		fmt.Printf("requested key=%q\n", key)
		var rspExists bool
		if rsp.Exists() != 0 {
			rspExists = true
		}
		fmt.Printf("responsed val=%q; exists=%t;\n", rsp.Val(), rspExists)

	case "put":
		reqKey := b.CreateString(key)
		reqVal := b.CreateString(val)

		keyval.PutReqStart(b)
		keyval.PutReqAddKey(b, reqKey)
		keyval.PutReqAddVal(b, reqVal)
		b.Finish(keyval.PutReqEnd(b))

		var rsp *keyval.PutRsp
		fmt.Printf("Sending Put(%q, %q)\n", key, val)
		rsp, err = client.Put(context.Background(), b)
		if err != nil {
			panic(errors.Wrap(err, "Put grpc request failed"))
		}

		fmt.Printf("requested Put key=%q val=%q\n", key, val)
		var rspAdded bool
		if rsp.Added() != 0 {
			rspAdded = true
		}
		fmt.Printf("responded added=%t\n", rspAdded)

	case "keys":
		panic(fmt.Errorf("%q cmd not implemented", cmd))
	default:
		panic(fmt.Errorf("unknown cmd: %q", cmd))
	}
	fmt.Println("done.")
}

func usage(out io.Writer, xit int, msg string) {
	fmt.Fprintln(out, msg)
	fmt.Fprintln(out, "go <cmd> [<cmd-arg>*]")
	fmt.Fprintln(out, "  ex#1 $ cli get \"key\"")
	fmt.Fprintln(out, "  ex#2 $ cli put \"key\" \"val\"")
	os.Exit(xit)
}
