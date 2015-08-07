package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

var apiPrefix = "/api/" + APIVersion

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) < len(apiPrefix) {
		log.Printf("Bad request URL: %s", r.URL.Path)
		http.Error(w, `{"error": "Bad request URL"}`, http.StatusNotFound)
		return
	}
	node := strings.TrimLeft(r.URL.Path[len(apiPrefix):], "/")

	var encoded []byte
	var err error

	vmap := safeVmap.Get()
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

// Reload and parse the virshFile periodically (runs as a goroutine)
func Reloader(refresh int) {
	var vmap Vmap
	for ;; {
		err := vmap.Load(virshFile)
		if err != nil {
			log.Printf("Problem getting vmap: %s", err.Error())
		}
		safeVmap.Set(vmap)
		log.Printf("Reloaded from %s, %d entries in map.\n", virshFile, vmap.Length())
		time.Sleep(time.Duration(refresh) * time.Minute)
	}
}

func Serve(address string) {
	http.HandleFunc("/api/v1/", handleRequest)
	log.Println("Starting server, listening on", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
