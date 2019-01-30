// Example key-value lookup/storage server.
// example runs:
// $ go run main.go -a "localhost:9090"
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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
	"github.com/pkg/errors"

	"google.golang.org/grpc"
	//"github.com/lleo/example-apps-go/key-val-network-app/keyval"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/lleo/example-apps-go/key-val-network-app/keyval"
)

type KeyVal = fmap.KeyVal
type server struct{}

var gblMapRWMu sync.RWMutex
var gblMap *fmap.Map

const defaultAddr = "0.0.0.0:9090"
const defaultVerbosity = false

var verbose bool

func init() {
	log.SetFlags(log.Lshortfile)

	var logFileName = "server.log"
	var logFile, err = os.Create(logFileName)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to os.Create(%q)", logFileName))
	}
	actionlog.SetOutput(logFile)
}

func main() {
	log.Println("len(os.Args) =", len(os.Args))
	log.Println("os.Args =", os.Args)
	log.Print("Starting...")

	var addr string
	flag.StringVar(&addr, "a", defaultAddr,
		"(hostname|ip):port (eg \"localhost:9090\") or \":9090\"")
	flag.BoolVar(&verbose, "v", defaultVerbosity, "set verbose logging")
	flag.Parse()

	var args = flag.Args()
	if len(args) != 0 {
		usage(os.Stdout, 0, "")
	}

	var svrSk, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	//gblMap = fmap.New()
	gblMap = fmap.NewFromList([]KeyVal{
		{Key: key.Str("foo"), Val: int32(1)},
		{Key: key.Str("bar"), Val: int32(1)},
		{Key: key.Str("baz"), Val: int32(1)},
	})

	log.Println("gblMap =", gblMap.String())

	grpcSvr := grpc.NewServer(grpc.CustomCodec(
		flatbuffers.FlatbuffersCodec{}))

	//var svr = new(server)
	keyval.RegisterKeyValSvcServer(grpcSvr, &server{})

	log.Println("started.")

	//register Service
	err = grpcSvr.Serve(svrSk)

	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func usage(out io.Writer, xit int, msg string) {
	fmt.Fprintln(out, msg)
	fmt.Fprintln(out, "svr [ -h | -a \"localhost:9090\" ]")
	flag.PrintDefaults()
	os.Exit(xit)
}

func (s *server) Store(
	ctx context.Context,
	req *keyval.StoreReq,
) (*flatbuffers.Builder, error) {
	_ = verbose && log.Printf("Store called: key=%q; val=%d\n", req.Key(), req.Val())

	var added bool
	gblMapRWMu.Lock()
	gblMap, added = gblMap.Store(key.Str(req.Key()), req.Val())
	gblMapRWMu.Unlock()

	_ = verbose && log.Printf("Sending StoreRsp(added=%t)\n", added)

	b := flatbuffers.NewBuilder(0)

	keyval.StoreRspStart(b)
	//Should be:
	//keyval.StoreRspAddAdded(b, added)
	if added {
		keyval.StoreRspAddAdded(b, 1)
	} else {
		keyval.StoreRspAddAdded(b, 0)
	}

	b.Finish(keyval.StoreRspEnd(b))

	return b, nil
}

func (s *server) Load(
	ctx context.Context,
	req *keyval.LoadReq,
) (*flatbuffers.Builder, error) {
	_ = verbose && log.Printf("Get called: key=%q\n", req.Key())

	gblMapRWMu.RLock()
	val, found := gblMap.Load(key.Str(req.Key()))
	gblMapRWMu.RUnlock()

	_ = verbose && log.Printf("type(val)=%T", val)
	_ = verbose && log.Printf("Sending LoadRsp(val=%v, found=%t)\n", val, found)

	b := flatbuffers.NewBuilder(0)

	var rspVal int32
	if found {
		rspVal = val.(int32)
	}

	keyval.LoadRspStart(b)
	keyval.LoadRspAddVal(b, rspVal)
	//should be:
	//keyval.LoadRspAddExists(b, found)
	if found {
		keyval.LoadRspAddExists(b, 1)
	} else {
		keyval.LoadRspAddExists(b, 0)
	}
	b.Finish(keyval.StoreRspEnd(b))

	return b, nil
}

func (s *server) Keys(
	ctx context.Context,
	req *keyval.KeysReq,
) (*flatbuffers.Builder, error) {
	_ = verbose && log.Println("Keys called:")

	_ = verbose && log.Println("gblMap.NumEntries() =", gblMap.NumEntries())

	gblMapRWMu.RLock()
	//var keys = make([]string, 0, gblMap.NumEntries())
	var keys []string
	gblMap.Range(func(kv fmap.KeyVal) bool {
		//_=verbose&&log.Println("kv =", kv)
		//k := string(kv.Key.(key.Str))
		//_=verbose&&log.Println("k =", k)
		//keys = append(keys, k)
		keys = append(keys, string(kv.Key.(key.Str)))
		return true
	})
	gblMapRWMu.RUnlock()

	_ = verbose && log.Println("len(keys) =", len(keys))
	_ = verbose && log.Println("keys =", keys)

	b := flatbuffers.NewBuilder(0)

	//keysVec := make([]flatbuffers.UOffsetT, 0, len(keys))
	var keysVec []flatbuffers.UOffsetT
	for _, key := range keys {
		keysVec = append(keysVec, b.CreateString(key))
	}

	_ = verbose && log.Println("len(keysVec) =", len(keysVec))

	var n = len(keys)
	keyval.KeysRspStartKeysVector(b, n)
	for i := n - 1; i >= 0; i-- {
		b.PrependUOffsetT(keysVec[i])
	}
	vec := b.EndVector(n)

	keyval.KeysRspStart(b)
	keyval.KeysRspAddKeys(b, vec)
	b.Finish(keyval.KeysRspEnd(b))

	return b, nil
}
