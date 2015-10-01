package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

var apiPrefix = "/api/" + APIVersion + "/"

type vmapHandler struct {
	mapCh chan *Vmap
}

// The HTTP handler for the API.  Returns results in JSON format.
func (v vmapHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) < len(apiPrefix) {
		log.Printf("Bad request URL: %s", r.URL.Path)
		http.Error(w, `{"error": "Bad request URL"}`, http.StatusNotFound)
		return
	}
	node := strings.TrimLeft(r.URL.Path[len(apiPrefix):], "/")
	var encoded []byte
	var err error

	vmap := <-v.mapCh
	if vmap.Length() == 0 {
		log.Printf("Vmap empty!")
		http.Error(w, `{"error": "Data source error"}`, http.StatusInternalServerError)
		return
	}
	if node == "" {
		log.Printf("Request for entire map, virtmap: %d nodes", vmap.Length())
		encoded, err = json.MarshalIndent(vmap, " ", "  ")
	} else {
		log.Printf("Request for %s, virtmap: %d nodes", node, vmap.Length())
		response, err := vmap.Get(node)
		if err != nil {
			encoded = []byte(`{"error": "Node ` + node + ` not found"}`)
		} else {
			encoded, err = json.MarshalIndent(response, " ", "  ")
		}
	}
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}"`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Server", "Virtmapper 0.0.3")
	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

// Registers the HTTP handler and runs the server.
func (v vmapHandler) Serve(address string) {
	http.HandleFunc(apiPrefix, v.handleRequest)
	log.Println("Starting server, listening on", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

// Reloads and parses the virshFile periodically
func Reloader(done <-chan struct{}, virshFile string, refresh int) chan *Vmap {
	var vmap *Vmap
	var delay time.Duration
	mapCh := make(chan *Vmap)
	go func() {
		for {
			select {
			case <-time.After(delay):
				err := vmap.Load(virshFile)
				if err != nil {
					log.Printf("Problem getting vmap: %s", err.Error())
				}
				log.Printf("Reloaded from %s, %d entries in map.\n", virshFile, vmap.Length())
				delay = time.Duration(refresh) * time.Minute
			case mapCh <- vmap:
			case <-done:
				log.Println("Reloader shutting down.")
				return
			}
		}
	}()
	return mapCh
}
