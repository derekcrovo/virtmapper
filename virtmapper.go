package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

// version constants
const (
	version    = "0.0.3"
	apiVersion = "v1"
)

// Configuration defaults
const (
	listenAddress     = ":7474"
	logFile           = "/var/log/virtmapper"
	refreshInterval   = 60 // Minutes
	ansibleOutputFile = "/tmp/virsh.txt"
)

func main() {
	app := cli.NewApp()
	app.Name = "virtmapper"
	app.Usage = "maps libvirt virtual guests to their hosts"
	app.Authors = []cli.Author{cli.Author{
		Name:  "Derek Crovo",
		Email: "dcrovo@gmail.com",
	}}
	app.Version = version
	cli.AppHelpTemplate = VirtmapperHelpTemplate
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
	}

	app.Commands = []cli.Command{
		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "run the server and accept map queries",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "address, a",
					Value: listenAddress,
					Usage: "address and port to listen on",
				},
				cli.StringFlag{
					Name:  "logfile, l",
					Value: logFile,
					Usage: "log file for server activity",
				},
				cli.IntFlag{
					Name:  "refreshInterval, r",
					Value: refreshInterval,
					Usage: "map refresh interval in minutes",
				},
				cli.StringFlag{
					Name:  "ansibleOutputFile, v",
					Value: ansibleOutputFile,
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
				v := server{Reloader(c.String("ansibleOutputFile"), c.Int("refreshInterval"))}
				v.Serve(c.String("address"))
			},
		},
		{
			Name:    "query",
			Aliases: []string{"q"},
			Usage:   "query a server with the given request",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "server, s",
					// Value: "manager.corp.airwave.com:7474",
					Usage: "address of server to query",
				},
			},
			Action: func(c *cli.Context) {
				result, err := Query(c.String("server"), c.Args().Get(0))
				if err != nil {
					os.Exit(1)
				}
				Display(result)
			},
		},
	}

	app.Run(os.Args)
}
