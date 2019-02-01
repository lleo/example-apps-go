//benchmark program
//ex: $ go run main.go -a "localhost:9090" -r 10 -w 3 -t 10min
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/lleo/example-apps-go/netapp-grpc-flatbuffers/keyval"
	"github.com/lleo/go-functional-collections/fmap"
	"github.com/pkg/errors"

	//"github.com/lleo/go-functional-collections/key"
	"github.com/lleo/stringutil"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
)

const defaultAddr = "localhost:9090"
const defaultReaders = 1 //10
const defaultWriters = 0 //3
const defaultVerbosity = false

type KeyVal = fmap.KeyVal

var KvsRwMu sync.RWMutex
var Kvs map[string]int32
var Keys []string

var verbose bool

func init() {
	log.SetFlags(log.Lshortfile)

	var logFileName = "bench.log"
	var logFile, err = os.Create(logFileName)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to os.Create(%q)", logFileName))
	}
	log.SetOutput(logFile)
}

var addr string
var numReaders int
var numWriters int

func main() {
	log.Println("len(os.Args) =", len(os.Args))
	log.Println("os.Args =", os.Args)

	//flag.StringVar(&addr, "a", defaultAddr,
	//	"(hostname|ip):port (eg \"localhost:9090\") or \":9090\"")
	//flag.IntVar(&numReaders, "r", defaultReaders, "100")
	//flag.IntVar(&numWriters, "w", defaultWriters, "10")
	//flag.Parse()

	app := &cli.App{} //cli.NewApp()

	app.Name = "bench"
	app.Usage = "bench [options]"
	app.UsageText = "bench [options]"
	app.HideVersion = false
	app.Version = "0.1.0"
	app.EnableShellCompletion = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "addr",
			Aliases:     []string{"a"},
			Usage:       "address of server host:port",
			Destination: &addr,
			Value:       defaultAddr,
		},
		&cli.IntFlag{
			Name:        "numReaders",
			Aliases:     []string{"n"},
			Usage:       "number of concurrent readers",
			Destination: &numReaders,
			Value:       defaultReaders,
		},
		&cli.IntFlag{
			Name:        "numWriters",
			Aliases:     []string{"w"},
			Usage:       "number of concurrent writers",
			Destination: &numWriters,
			Value:       defaultWriters,
		},
		&cli.BoolFlag{
			Name:        "verbose",
			Aliases:     []string{"v"},
			Usage:       "set verbose logging",
			Destination: &verbose,
			Value:       defaultVerbosity,
		},
	}

	app.Action = doit
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}

	os.Exit(0)
}

func doit(c *cli.Context) error {
	log.Printf("addr = %q\n", addr)
	log.Printf("numReader = %d\n", numReaders)
	log.Printf("numWriters = %d\n", numWriters)

	client := startClient(addr)

	//
	// build key/val data
	//
	KvsRwMu.Lock()
	//Kvs, Keys = buildKvs(5)
	Kvs, Keys = getKvs(client)
	KvsRwMu.Unlock()

	//
	//Spawn Goroutines
	//
	var wg sync.WaitGroup
	var exit = time.After(30 * time.Second)

	//spawn readers
	for i := 0; i < numReaders; i++ {
		go spawnReader(client, i, &wg, exit)
	}

	//spawn writers
	for i := 0; i < numWriters; i++ {
		go spawnWriter(client, i, &wg, exit)
	}

	log.Println("Waiting for goroutines to end.")
	wg.Wait()
	log.Println("The End.")

	return nil
}

var Inc = stringutil.Lower.Inc

func startClient(addr string) *keyval.KeyValSvcClient {
	//
	// build KeyValSvcClient <- that is an interface
	//
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

	return client
}

func cmdKeys(client keyval.KeyValSvcClient) *keyval.KeysRsp {
	b := flatbuffers.NewBuilder(0)

	keyval.KeysReqStart(b)
	b.Finish(keyval.KeysReqEnd(b))

	var rsp *keyval.KeysRsp
	var err error
	log.Println("Sending Keys()")
	rsp, err = client.Keys(context.Background(), b)
	if err != nil {
		panic(errors.Wrap(err, "Keys grpc request failed"))
	}

	return rsp
}

func getKeys(client keyval.KeyValSvcClient) []string {
	var rsp = cmdKeys(client)

	keys := make([]string, 0, rsp.KeysLength())
	log.Printf("#keys = %d\n", rsp.KeysLength())
	for i := 0; i < rsp.KeysLength(); i++ {
		log.Printf("[%d] %q\n", i, rsp.Keys(i))
		keys = append(keys, string(rsp.Keys(i)))
	}

	return keys
}

func getKvs(client keyval.KeyValSvcClient) (map[string]int32, []string) {
	keys := getKeys(client)

	kvs := make(map[string]int32)
	for _, k := range keys {
		rsp := cmdLoad(client, k)
		kvs[k] = rsp.Val()
	}

	return kvs, keys
}

func buildKvs(num int) (map[string]int32, []string) {
	var kvs = make(map[string]int32, num)
	var keys = make([]string, 0, num)

	for i, key := int32(0), "aaa"; i < int32(num); i, key = i+1, Inc(key) {
		kvs[key] = 1
		keys = append(keys, key)
	}

	return kvs, keys
}

func spawnReader(
	client keyval.KeyValSvcClient,
	i int,
	wg *sync.WaitGroup,
	exit <-chan time.Time,
) {
	wg.Add(1)

LOOP:
	for /*ever*/ {
		select {
		case <-exit:
			break LOOP
		default:
			KvsRwMu.RLock()
			key := randomKey(Keys)
			KvsRwMu.RUnlock()
			cmdLoad(client, key)
		}
	}
	wg.Done()
}

func randomKey(keys []string) string {
	return keys[rand.Intn(len(keys))]
}

func cmdLoad(client keyval.KeyValSvcClient, key string) *keyval.LoadRsp {
	b := flatbuffers.NewBuilder(0)

	reqKey := b.CreateString(key)

	keyval.LoadReqStart(b)
	keyval.LoadReqAddKey(b, reqKey)
	b.Finish(keyval.LoadReqEnd(b))

	rsp, err := client.Load(context.Background(), b)
	if err != nil {
		log.Fatalf("client.Load failed: %v", err)
	}

	//log.Printf("LoadRsp.val: %T(%v)\n", rsp.Val(), rsp.Val())
	//log.Printf("LoadRsp.exists: %T(%v)\n", rsp.Exists(), rsp.Exists())

	return rsp
}

func spawnWriter(
	client keyval.KeyValSvcClient,
	i int,
	wg *sync.WaitGroup,
	exit <-chan time.Time,
) {
	wg.Add(1)

LOOP:
	for /*ever*/ {
		select {
		case <-exit:
			break LOOP
		default:
			numKeys := len(Keys)
			n := rand.Intn(numKeys + 1)
			var k string
			KvsRwMu.Lock() // Write Lock Kvs & Keys
			if n == numKeys {
				k = Inc(Keys[numKeys-1])
				Keys = append(Keys, k)
				Kvs[k] = 1
				numKeys++
			} else {
				k = Keys[n]
				Kvs[k]++
			}
			v := Kvs[k]
			KvsRwMu.Unlock() // Write Unlock Kvs & Keys
			cmdStore(client, k, v)
		}
	}
	wg.Done()
}

func cmdStore(client keyval.KeyValSvcClient, k string, v int32) *keyval.StoreRsp {
	b := flatbuffers.NewBuilder(0)

	reqKey := b.CreateString(k)

	keyval.StoreReqStart(b)
	keyval.StoreReqAddKey(b, reqKey)
	keyval.StoreReqAddVal(b, v)
	b.Finish(keyval.StoreReqEnd(b))

	var rsp *keyval.StoreRsp
	var err error
	//log.Printf("Sending Store(%q, %d)\n", k, v)
	rsp, err = client.Store(context.Background(), b)
	if err != nil {
		panic(errors.Wrap(err, "Store grpc request failed"))
	}

	//log.Printf("requested Store key=%q val=%d\n", k, v)
	//log.Printf("responded added=%T(%v)\n", rsp.Added(), rsp.Added())

	return rsp
}
