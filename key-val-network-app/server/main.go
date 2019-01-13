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
	"sync"

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

type server struct{}

var gblMapRWMu sync.RWMutex
var gblMap *fmap.Map

func (s *server) Put(ctx context.Context, req *keyval.PutReq) (*flatbuffers.Builder, error) {
	log.Printf("Put called: key=%q; val=%q\n", req.Key(), req.Val())

	var added bool
	var gblMap0 *fmap.Map
	gblMapRWMu.Lock()
	gblMap0, added = gblMap.Store(key.Str(req.Key()), string(req.Val()))
	gblMap = gblMap0
	gblMapRWMu.Unlock()

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

func (s *server) Get(ctx context.Context, req *keyval.GetReq) (*flatbuffers.Builder, error) {
	log.Printf("Get called: key=%q\n", req.Key())

	gblMapRWMu.RLock()
	val, found := gblMap.Load(key.Str(req.Key()))
	gblMapRWMu.RUnlock()

	log.Printf("type(val)=%T", val)
	log.Printf("Sending GetRsp(val=%q, found=%t)\n", val, found)

	b := flatbuffers.NewBuilder(0)

	var rspVal flatbuffers.UOffsetT
	if val != nil {
		rspVal = b.CreateString(val.(string))
	} else {
		rspVal = b.CreateString("")
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

	gblMap = fmap.New()

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
