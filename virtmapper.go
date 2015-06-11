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
	httpServer   string
	query        string
	printVersion bool
)

func init() {
	flag.StringVar(&httpAddr, "http", "", "HTTP service address (e.g., ':6060')")
	flag.StringVar(&httpServer, "server", "", "HTTP server to query (e.g., 'server.example.com:6060')")
	flag.StringVar(&query, "query", "", "Host to query about, omit for all hosts")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
}

func main() {
	flag.Parse()

	if printVersion {
		fmt.Printf("virtmapper %s\n", Version)
		os.Exit(0)
	}

	if httpAddr != "" && httpServer != "" {
		fmt.Println("Please specify either -server or -http, not both")
		os.Exit(1)
	}

	if httpAddr != "" {
		Serve()
	}

	if httpServer != "" {
		Display(Query(query))
//		fmt.Printf("Query returned %s", r)
	}
}
