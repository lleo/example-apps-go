// Example key-value lookup/storage client.
// example runs:
// $ go run main.go -a "localhost:9090" put foo '{"json": true}'
// $ go run main.go get foo
// $ go run main.go keys fee fie foe fum //returns key/val pairs for key that exist
// $ go run main.go keys #gets all key/val pairs
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"context"

	"github.com/pkg/errors"
	flatbuffers "github.com/google/flatbuffers/go"
	//"github.com/lleo/example-apps-go/key-val-network-app/keyval"
	"github.com/lleo/example-apps-go/key-val-network-app/keyval"
	"google.golang.org/grpc"
)

 const defaultPort = "9090"
const defaultIP = "localhost"

func usage(out io.Writer, xit int, msg string) {
	fmt.Fprintln(out, msg)
	fmt.Fprintln(out, "go <cmd> [<cmd-arg>*]")
	fmt.Fprintln(out, "  ex#1 $ cli get \"key\"")
	fmt.Fprintln(out, "  ex#2 $ cli put \"key\" \"val\"")
	os.Exit(xit)
}

func main() {
	fmt.Println("len(os.Args) =", len(os.Args))
	fmt.Println("os.Args =", os.Args)

	var addr string
	//flags.StringVar(&addr, "a", defaultIp+":"+defaultPort,
	//	"(hostname|ip):port (eg \"localhost:9090\")")
	addr = defaultIP + ":" + defaultPort

	if len(os.Args) < 2 {
		usage(os.Stdout, 0, "")
	}

	b := flatbuffers.NewBuilder(0)

	cmd := strings.ToLower(os.Args[1])
	var key string
	//var val string
	switch cmd {
	case "get":
		if len(os.Args) != 3 {
			usage(os.Stderr, 1, "args != 3")
		}
		key=os.Args[2] // $ cli[0] get[1] "value"[2]
	case "put":
		usage(os.Stderr, 1, "\"put\" not implemented.")
	case "getkeys":
		usage(os.Stderr, 1, "\"getkeys\" not implemented.")
	default:
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
		var rsp *keyval.GetRsp;
		fmt.Printf("Sending Get(%q)\n", key)
		rsp, err = client.Get(context.Background(), b)
		if err != nil {
			panic(errors.Wrap(err, "Get grpc request failed"))
		}

		fmt.Printf("requested key=%q\n", key)
		fmt.Printf("responsed val=%q\n", rsp.Val())
	case "put":
		panic(fmt.Errorf("%q cmd not implemented", cmd))
		//var rsp *keyval.PutRsp
		//rsp, err = client.Put(context.Background(), b)
	case "keys":
		panic(fmt.Errorf("%q cmd not implemented", cmd))
	default:
		panic(fmt.Errorf("unknown cmd: %q", cmd))
	}
	fmt.Println("done.")
}
