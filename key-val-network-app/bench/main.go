//benchmark program
//ex: $ go run main.go -a "localhost:9090" -r 10 -w 3 -t 10min
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const defaultAddr = "localhost:9090"
const defaultReaders = 10
const defaultWriters = 3

func main() {
	fmt.Println("len(os.Args) =", len(os.Args))
	fmt.Println("os.Args =", os.Args)

	var addr string
	var numReaders int
	var numWriters int

	flag.StringVar(&addr, "a", defaultAddr,
		"(hostname|ip):port (eg \"localhost:9090\") or \":9090\"")
	flag.IntVar(&numReaders, "r", defaultReaders, "100")
	flag.IntVar(&numWriters, "w", defaultWriters, "10")
	flag.Parse()

	var args = flag.Args()
	if len(args) > 0 {
		usage(os.Stdout, 0, "")
	}

	fmt.Printf("addr = %q\n", addr)
	fmt.Printf("numReader = %d\n", numReaders)
	fmt.Printf("numWriters = %d\n", numWriters)
}

func usage(out io.Writer, xit int, msg string) {
	fmt.Fprintln(out, msg)
	fmt.Fprintln(out, "go <cmd> [<cmd-arg>*]")
	fmt.Fprintln(out, "  ex#1 $ cli get \"key\"")
	fmt.Fprintln(out, "  ex#2 $ cli put \"key\" \"val\"")
	os.Exit(xit)
}
