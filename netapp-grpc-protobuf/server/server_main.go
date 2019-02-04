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
	"google.golang.org/grpc"
)

var _ = fmt.Printf
var _ = log.Printf

var addr = ":9090"

type server struct{}

type KeyVal = fmap.KeyVal

var KvsRwMu sync.RWMutex
var Kvs map[string]int32

func main() {
	flag.StringVar(&addr, "a", ":9090", "<host:port> string")

	KvsRwMu.Lock()
	Kvs = make(map[string]int32)
	Kvs["foo"] = 1
	Kvs["bar"] = 1
	Kvs["baz"] = 1
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

	KvsRwMu.RLock()
	v, exists := Kvs[req.Key]
	KvsRwMu.RUnlock()

	var rsp *keyval.LoadRsp
	rsp = &keyval.LoadRsp{Val: v, Exists: exists}
	return rsp, nil
}

func (s *server) Store(
	ctx context.Context,
	req *keyval.StoreReq,
) (*keyval.StoreRsp, error) {
	log.Println("Store() grpc called.")

	KvsRwMu.Lock()
	_, exists := Kvs[req.Key]
	Kvs[req.Key] = req.Val
	KvsRwMu.Unlock()

	var rsp *keyval.StoreRsp
	rsp = &keyval.StoreRsp{Added: !exists}
	return rsp, nil
}

func (s *server) Keys(
	ctx context.Context,
	req *keyval.KeysReq,
) (*keyval.KeysRsp, error) {
	log.Println("Keys() grpc called.")

	var keys []string
	KvsRwMu.Lock()
	for k, _ := range Kvs {
		keys = append(keys, k)
	}
	KvsRwMu.Unlock()

	rsp := &keyval.KeysRsp{Keys: keys}

	return rsp, nil
}
