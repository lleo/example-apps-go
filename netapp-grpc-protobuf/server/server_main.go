package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/lleo/example-apps-go/netapp-grpc-protobuf/keyval"
	"github.com/lleo/go-functional-collections/fmap"
	"github.com/lleo/go-functional-collections/key"
	"google.golang.org/grpc"
)

var _ = fmt.Printf
var _ = log.Printf

var addr = ":9090"

type server struct{}

type KeyVal = fmap.KeyVal

var KvsRwMu sync.RWMutex
var Kvs []KeyVal
var Keys []string

func main() {
	flag.StringVar(&addr, "a", ":9090", "<host:port> string")

	KvsRwMu.Lock()
	Kvs = []KeyVal{
		KeyVal{Key: key.Str("foo"), Val: int32(1)},
		KeyVal{Key: key.Str("bar"), Val: int32(1)},
		KeyVal{Key: key.Str("baz"), Val: int32(1)},
	}
	KvsRwMu.Unlock()

	svrSk, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to open listen socket: %v", err)
	}
	svr := grpc.NewServer()
	keyval.RegisterKeyValSvcServer(svr, &server{})
	err = svr.Serve(svrSk)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) Load(
	ctx context.Context,
	req *keyval.LoadReq,
) (*keyval.LoadRsp, error) {
	log.Println("Load() grpc called.")
	return nil, nil
}

func (s *server) Store(
	ctx context.Context,
	req *keyval.StoreReq,
) (*keyval.StoreRsp, error) {
	log.Println("Store() grpc called.")
	return nil, nil
}

func (s *server) Keys(
	ctx context.Context,
	req *keyval.KeysReq,
) (*keyval.KeysRsp, error) {
	log.Println("Keys() grpc called.")

	KvsRwMu.Lock()
	for _, kv := range Kvs {
		Keys = append(Keys, string(kv.Key.(key.Str)))
	}
	KvsRwMu.Unlock()

	rsp := &keyval.KeysRsp{Keys: Keys}

	return rsp, nil
}
