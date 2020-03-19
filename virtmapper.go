package main

import (
	"fmt"
	"os"
)

// Version constants
const (
	Version    = "0.0.5"
	APIVersion = "v1"
)

// Configuration defaults
const (
	ListenAddress     = ":7474"
	LogFile           = "/var/log/virtmapper"
	RefreshInterval   = 60 // Minutes
	AnsibleOutputFile = "/tmp/virtmapper.txt"
)

func main() {
	app := CLIApp()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("App error: %v\n", err)
	}
}
