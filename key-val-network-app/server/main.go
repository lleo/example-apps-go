// Example key-value lookup/storage server.
// example runs:
// $ go run main.go -a "localhost:9090"
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	//"flags"
	//"gopkg.in/urfave/cli.v2" // imports as package "cli"
	//"github.com/lleo/go-functional-collections/fmap"
	//"github.com/lleo/go-functional-collections/key"
	"github.com/lleo/go-functional-collections/fmap"
	"github.com/lleo/go-functional-collections/key"

	"google.golang.org/grpc"
	//"github.com/lleo/example-apps-go/key-val-network-app/keyval"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/lleo/example-apps-go/key-val-network-app/keyval"
)

const defaultPort = 6060
const defaultIP = "localhost"

type server struct{}

var m *fmap.Map

func (s *server) Put(
	ctx context.Context,
	req *keyval.PutReq) (*flatbuffers.Builder, error) {
	log.Printf("Put called: key=%q; val=%q\n",
		req.Key(), req.Val())

	var added bool
	m, added = m.Store(key.Str(req.Key()), req.Val())

	log.Printf("Sending PutRsp(added=%t)\n", added)

	b := flatbuffers.NewBuilder(0)

	keyval.PutRspStart(b)
	//Should be:
	//keyval.PutRspAddAdded(b, added)
	if added {
		keyval.PutRspAddAdded(b, 1)
	} else {
		keyval.PutRspAddAdded(b, 0)
	}

	b.Finish(keyval.PutReqEnd(b))

	return b, nil
}

func (s *server) Get(ctx context.Context,
	req *keyval.GetReq) (*flatbuffers.Builder, error) {
	log.Printf("Get called: key=%q\n", req.Key())

	val, found := m.Load(key.Str(req.Key()))

	log.Printf("Sending GetRsp(val=%q, found=%t)\n", val, found)

	b := flatbuffers.NewBuilder(0)
	var rspVal flatbuffers.UOffsetT
	if val != nil {
		rspVal = b.CreateString(val.(string))
	} else {
		rspVal = b.CreateString("empty")
	}

	keyval.GetRspStart(b)
	keyval.GetRspAddVal(b, rspVal)
	//should be:
	//keyval.GetRspAddExists(b, found)
	if found {
		keyval.GetRspAddExists(b, 1)
	} else {
		keyval.GetRspAddExists(b, 0)
	}
	b.Finish(keyval.PutRspEnd(b))

	return b, nil
}

func main() {
	fmt.Println("len(os.Args) =", len(os.Args))
	fmt.Println("os.Args =", os.Args)
	fmt.Print("Starting...")

	addr := "0.0.0.0:9090"
	var svrSk, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	m = fmap.New()

	grpcSvr := grpc.NewServer(grpc.CustomCodec(
		flatbuffers.FlatbuffersCodec{}))

	keyval.RegisterKeyValSvcServer(grpcSvr, &server{})

	fmt.Println("started.")

	//register Service
	err = grpcSvr.Serve(svrSk)

	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
