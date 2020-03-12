package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli"
)

// Authors is the cli authors struct for the help message
var Authors = []cli.Author{{
	Name: "Derek Crovo",
}}

// HTTPGetter is a helper func for http.Get(), factored out so it can be a hook for testing.
var HTTPGetter = func(url string) (*http.Response, error) {
	return http.Get(url)
}

// Query is the cli client function.  It queries the given server for the
// given host, unmarshalls the JSON, and returns a result Vmap pointer.
func Query(httpServer string, query string) (*Vmap, error) {
	vmap := &Vmap{}
	rawResponse, err := HTTPGetter("http://" + httpServer + APIPrefix + "vmap/" + query)
	if err != nil {
		fmt.Printf("Get() error, %v\n", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		fmt.Printf("ReadAll() error, %v\n", err)
		return nil, err
	}

	var decodedResponse map[string]interface{}
	err = json.Unmarshal(body, &decodedResponse)
	if err != nil {
		fmt.Printf("JSON Unmarshalling error: %v\n", err)
		return nil, err
	}

	// If the response contains a top-level error, return it
	if data, ok := decodedResponse["error"]; ok {
		return nil, errors.New(data.(string))
	}

	err = json.Unmarshal(body, vmap)
	if err != nil {
		fmt.Printf("Unmarshal() error, %v\n", err)
		return nil, err
	}
	return vmap, nil
}

// Display takes a result Vmap from Query() and displays it to the user.
func Display(vmap *Vmap) {
	for n := range vmap.Hosts {
		fmt.Println(vmap.Info(n))
	}
	for n := range vmap.Guests {
		fmt.Println(vmap.Info(n))
	}
}

// CLIApp creates the cli application with commands and config defaults
func CLIApp() *cli.App {
	app := cli.NewApp()
	app.Name = "virtmapper"
	app.Usage = "maps libvirt virtual guests to their hosts"
	app.Authors = Authors
	app.Version = Version
	cli.AppHelpTemplate = CLIHelpTemplate
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
	}

	app.Commands = []cli.Command{{
		Name:    "serve",
		Aliases: []string{"s"},
		Usage:   "run the server and accept map queries",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "address, a",
				Value: ListenAddress,
				Usage: "address and port to listen on",
			},
			cli.StringFlag{
				Name:  "logfile, l",
				Value: LogFile,
				Usage: "log file for server activity",
			},
			cli.IntFlag{
				Name:  "refreshInterval, r",
				Value: RefreshInterval,
				Usage: "map refresh interval in minutes",
			},
			cli.StringFlag{
				Name:  "ansibleOutputFile, v",
				Value: AnsibleOutputFile,
				Usage: "path to Ansible output file to read",
			},
		},
		Action: func(c *cli.Context) {
			f, err := os.OpenFile(c.String("logfile"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				fmt.Printf("Error opening file: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			log.SetOutput(f)
			v := newServer()
			v.Serve(c)
		},
	}, {
		Name:    "query",
		Aliases: []string{"q"},
		Usage:   "query a server with the given request",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "server, s",
				Usage: "address of server to query",
				Value: "localhost:7474",
			},
		},
		Action: func(c *cli.Context) {
			result, err := Query(c.String("server"), c.Args().Get(0))
			if err != nil {
				fmt.Printf("Query error: %v\n", err)
				os.Exit(1)
			}
			Display(result)
		},
	}}
	return app
}

// CLIHelpTemplate is the CLI help text template
var CLIHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}{{if len .Authors}}

AUTHOR:
   {{range .Authors}}{{ . }}{{end}}{{end}}

COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}{{end}}
`
