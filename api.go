package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const apiPrefix = "/api/" + apiVersion + "/"

type server struct {
	mapCh chan *Vmap
}

// The HTTP handler for the API.  Returns results in JSON format.
func (s server) handleRequest(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) < len(apiPrefix) {
		err := fmt.Errorf("Bad request URL: %s", r.URL.Path)
		log.Println(err)
		s.responderr(w, r, http.StatusNotFound, err)
		return
	}
	node := strings.TrimLeft(r.URL.Path[len(apiPrefix):], "/")
	vmap := <-s.mapCh
	if vmap.Length() == 0 {
		log.Printf("Vmap is empty")
	}
	var response Vmap
	if node == "" {
		log.Printf("Request for entire map, virtmap: %d nodes", vmap.Length())
		response = *vmap
	} else {
		log.Printf("Request for %s, virtmap: %d nodes", node, vmap.Length())
		var err error
		response, err = vmap.Get(node)
		if err == errNodeNotFound {
			s.responderr(w, r, http.StatusNotFound, fmt.Errorf("Node %s not found", node))
			return
		}
		if err != nil {
			s.responderr(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	s.respond(w, r, http.StatusOK, response)
}

func (s server) respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Server", "Virtmapper v"+version)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("encode response: %s", err)
	}
}

func (s server) responderr(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.Header().Set("Server", "Virtmapper v"+version)
	w.WriteHeader(status)
	var data struct {
		Error string `json:"error"`
	}
	if err != nil {
		data.Error = err.Error()
	} else {
		data.Error = "Something went wrong"
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("encode response: %s", err)
	}
}

// Registers the HTTP handler and runs the server.
func (s server) Serve(address string) {
	http.HandleFunc(apiPrefix, s.handleRequest)
	log.Println("Starting server, listening on", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

// Reloader launches a goroutine which reloads and parses the virshFile periodically
func Reloader(virshFile string, refresh int) chan *Vmap {
	vmap := new(Vmap)
	var delay time.Duration
	mapCh := make(chan *Vmap)
	go func() {
		for {
			select {
			case <-time.After(delay):
				vmap = &Vmap{}
				err := vmap.Load(virshFile)
				if err != nil {
					log.Printf("Problem getting vmap: %s", err.Error())
				}
				log.Printf("Reloaded from %s, %d entries in map.\n", virshFile, vmap.Length())
				delay = time.Duration(refresh) * time.Minute
			case mapCh <- vmap:
			}
		}
	}()
	return mapCh
}
