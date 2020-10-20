package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/datewu/sandy"
)

var (
	addr             = flag.String("addr", ":1200", "server udp address")
	serverMode       = flag.Bool("s", false, "runing mode, default as client mode")
	serverCpuprofile = flag.String("scpu", "", "write cpu profile to file")
	cliCpuprofile    = flag.String("ccpu", "", "write cpu profile to file")
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
	if *serverCpuprofile != "" {
		f, err := os.Create(*serverCpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		go func() {
			time.Sleep(10 * time.Minute)
			pprof.StopCPUProfile()
		}()
	}
	face := func(fileName string, id string) (io.WriteCloser, error) {
		return os.Create(fileName + "." + id + ".debug")
	}
	sandy.Serve(*addr, face)
}

func client() {
	if *cliCpuprofile != "" {
		f, err := os.Create(*cliCpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if len(flag.Args()) != 1 {
		panic("no file; no peanut; no protein")
	}
	// one peanut a time, sandy.
	// only one peanut for test
	fName := flag.Args()[0]
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
	go func() {
		err := sandy.Upload(*addr, p)
		if err != nil {
			log.Println("upload failed", err)
		}
	}()
	for msg := range p.Feedback {
		fmt.Print(msg)
	}
}
