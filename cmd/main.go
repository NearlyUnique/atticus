package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/NearlyUnique/atticus"
)

func main() {
	control := flag.String("control", ":10000", "control plane listen interface")
	run := flag.String("runtime", ":10001", "runtime listen interface")
	initial := flag.String("initial", "", "initial canned responses")

	flag.Parse()

	runListener, err := net.Listen("tcp", *run)
	if err != nil {
		log.Printf("Failed to bind to (run) %s: %v", *run, err)
		os.Exit(1)
		return
	}
	ctrlListener, err := net.Listen("tcp", *control)
	if err != nil {
		log.Printf("Failed to bind to (control) %s: %v", *run, err)
		os.Exit(1)
		return
	}

	s, err := atticus.New(*initial)

	if err != nil {
		log.Printf("Failed to initialise to %s: %v", *initial, err)
		os.Exit(1)
		return
	}

	err = s.Run(ctrlListener, runListener)

	if err != nil {
		log.Printf("terminated:%v", err)
	}
}
