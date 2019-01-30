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
	"log"
	"os"
	"strconv"
	"strings"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/pkg/errors"

	//"github.com/lleo/example-apps-go/key-val-network-app/keyval"
	"github.com/lleo/example-apps-go/key-val-network-app/keyval"
	"google.golang.org/grpc"
)

const defaultAddr = "localhost:9090"

func init() {
	log.SetFlags(log.Lshortfile)

	var logFileName = "client.log"
	var logFile, err = os.Create(logFileName)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to os.Create(%q)", logFileName))
	}
	log.SetOutput(logFile)
}

func main() {
	log.Println("len(os.Args) =", len(os.Args))
	log.Println("os.Args =", os.Args)

	var addr string
	flag.StringVar(&addr, "a", defaultAddr,
		"(hostname|ip):port (eg \"localhost:9090\") or \":9090\"")
	flag.Parse()

	var args = flag.Args()
	log.Printf("len(args)=%d\n", len(args))
	if len(args) < 1 {
		usage(os.Stdout, 0, "requires a command load/store/keys")
	}

	b := flatbuffers.NewBuilder(0)

	cmd := strings.ToLower(args[0])
	var key string
	var val64 int64
	var val int32
	var err error
	switch cmd {
	case "load":
		if len(args) != 2 {
			usage(os.Stderr, 1, "args != 2")
		}
		key = args[1] // $ cli -a ":9090" get[0] "value"[1]
		log.Println("command:", cmd, key)
	case "store":
		if len(args) != 3 {
			usage(os.Stderr, 1, "args != 3")
		}
		key = args[1]
		val64, err = strconv.ParseInt(args[2], 10, 32)
		if err != nil {
			panic(err)
		}
		val = int32(val64)
		log.Println("command:", cmd, key, val)
	case "keys":
		if len(args) != 1 {
			usage(os.Stderr, 1, "args != 1")
		}
		log.Println("command:", cmd)
	default:
		usage(os.Stderr, 1, "unknown command "+cmd)
	}

	var conn *grpc.ClientConn
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
	case "load":
		//build request

		//must do b.CreateString() outside of keyval.*Start() and
		//keyval.*Finish() calls or else I get "Incorrect creation order:
		//object must /not be nested." error panic from keyval library.
		reqKey := b.CreateString(key)

		keyval.LoadReqStart(b)
		keyval.LoadReqAddKey(b, reqKey)
		b.Finish(keyval.LoadReqEnd(b))

		//client call
		var rsp *keyval.LoadRsp
		log.Printf("Sending Load(%q)\n", key)
		rsp, err = client.Load(context.Background(), b)
		if err != nil {
			panic(errors.Wrap(err, "Load grpc request failed"))
		}

		log.Printf("requested key=%q\n", key)
		log.Printf("responsed val=%T(%v); exists=%T(%v);\n",
			rsp.Val(), rsp.Val(), rsp.Exists(), rsp.Exists())

	case "store":
		reqKey := b.CreateString(key)

		keyval.StoreReqStart(b)
		keyval.StoreReqAddKey(b, reqKey)
		keyval.StoreReqAddVal(b, val)
		b.Finish(keyval.StoreReqEnd(b))

		var rsp *keyval.StoreRsp
		log.Printf("Sending Store(%q, %d)\n", key, val)
		rsp, err = client.Store(context.Background(), b)
		if err != nil {
			panic(errors.Wrap(err, "Store grpc request failed"))
		}

		log.Printf("requested Store key=%q val=%d\n", key, val)
		log.Printf("responded added=%T(%v)\n", rsp.Added(), rsp.Added())

	case "keys":
		keyval.KeysReqStart(b)
		b.Finish(keyval.KeysReqEnd(b))

		var rsp *keyval.KeysRsp
		log.Println("Sending Keys()")
		rsp, err = client.Keys(context.Background(), b)
		if err != nil {
			panic(errors.Wrap(err, "Keys grpc request failed"))
		}
		log.Printf("#keys = %d\n", rsp.KeysLength())
		for i := 0; i < rsp.KeysLength(); i++ {
			log.Printf("[%d] %q\n", i, rsp.Keys(i))
		}
	default:
		panic(fmt.Errorf("unknown cmd: %q", cmd))
	}
	log.Println("done.")
}

func usage(out io.Writer, xit int, msg string) {
	fmt.Fprintln(out, msg)
	fmt.Fprintln(out, "go <cmd> [<cmd-arg>*]")
	fmt.Fprintln(out, "  ex#1 $ cli load \"key\"")
	fmt.Fprintln(out, "  ex#2 $ cli store \"key\" <val>")
	fmt.Fprintln(out, "  ex#3 $ cli keys")
	os.Exit(xit)
}
