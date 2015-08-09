package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/codegangsta/cli"
)

const Version = "0.0.3"
const APIVersion = "v1"

// SafeVmp is a mutex-protected vmap struct
type SafeVmap struct {
	vmap   Vmap
	rwlock sync.RWMutex
}

// Global vmap which is used for queries
// and set by the Reloader function
var safeVmap SafeVmap

func (s *SafeVmap) Get() Vmap {
	s.rwlock.RLock()
	defer s.rwlock.RUnlock()
	return s.vmap
}

func (s *SafeVmap) Set(v Vmap) {
	s.rwlock.Lock()
	defer s.rwlock.Unlock()
	s.vmap = v
}

func main() {
	app := cli.NewApp()
	app.Name = "virtmapper"
	app.Usage = "maps virtual guests to their hosts"
	app.Version = "0.0.3"
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
					Value: ":7474",
					Usage: "address and port to listen on",
				},
				cli.StringFlag{
					Name:  "logfile, l",
					Value: "/var/log/virtmapper",
					Usage: "log file for server activity",
				},
				cli.IntFlag{
					Name:  "reload, r",
					Value: 60,
					Usage: "map refresh interval in minutes",
				},
				cli.StringFlag{
					Name:  "virshfile, v",
					Value: "/tmp/virsh.txt",
					Usage: "path to virsh dump file to read",
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
				go Reloader(c.String("virshfile"), c.Int("reload"))
				Serve(c.String("address"))
			},
		},
		{
			Name:    "query",
			Aliases: []string{"q"},
			Usage:   "query a server with the given request",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server, s",
					Value: "manager.corp.airwave.com:7474",
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
