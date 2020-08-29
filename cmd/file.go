package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/datewu/sandy"
)

var (
	addr       = flag.String("addr", ":1200", "server udp address")
	serverMode = flag.Bool("s", false, "runing mode, default as client mode")
)

func main() {
	flag.Parse()
	fmt.Println("Hi, SpongeBob SquarePants!")

	if *serverMode {
		server()
		return
	}
	client()
}

func server() {
	face := func(fileName string, id string) (io.WriteCloser, error) {
		return os.Create(fileName + "." + id + ".debug")
	}
	sandy.Serve(*addr, face)
}

func client() {
	if len(os.Args) < 2 {
		panic("no file; no peanut; no protein")
	}
	// one peanut a time, sandy.
	// only one peanut for test
	fName := os.Args[1]
	f, err := os.Open(fName)
	if err != nil {
		panic(err)
	}
	info, err := f.Stat()
	if err != nil {
		panic(err)
	}

	p := &sandy.Peanut{
		Protein:  f,
		Name:     fName,
		Size:     info.Size(),
		Feedback: make(chan string),
	}
	go sandy.Upload(*addr, p)
	for msg := range p.Feedback {
		fmt.Print(msg)
	}
}
