package main

import (
	"flag"
	"fmt"
	"os"
)

const Version = "0.0.1"
const virsh_file = "/tmp/virsh.txt"

var (
	httpAddr     string
	printVersion bool
)

func init() {
	flag.StringVar(&httpAddr, "http", "", "HTTP service address (e.g., ':6060')")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
}

func main() {
	flag.Parse()

	if printVersion {
		fmt.Printf("virtmapper %s\n", Version)
		os.Exit(0)
	}

	if httpAddr != "" {
		Serve()
	}
}
