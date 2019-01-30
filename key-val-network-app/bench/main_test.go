package main

import (
	"os"
	"testing"
	//flatbuffers "github.com/google/flatbuffers/go"
	//"github.com/lleo/example-apps-go/key-val-network-app/keyval"
)

var testAddr = "localhost:9090"

func TestMain(m testing.M) {

	//FIXME: urfave.cli stuff here

	os.Exit(m.Run())
}

func BenchmarkBasicLoad(b *testing.B) {
	client := startClient(testAddr)

	for i := 0; i < b.N; i++ {
		KvsRwMu.RLock()
		_ = randomeKey(Keys)
		KvsRwMu.RUnlock()
	}
}

func BenchmarkCmdLoad(b *testing.B) {
	client := startClient(testAddr)

	for i := 0; i < b.N; i++ {
		KvsRwMu.RLock()
		_ = randomeKey(Keys)
		KvsRwMu.RUnlock()

		_ = cmdLoad(client)
	}
}
