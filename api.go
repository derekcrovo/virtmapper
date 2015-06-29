package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/subsonic74/virtmapper/virtmap"
)

const api_prefix = "/api/v1"

var GetNodes = func() (virtmap.Vmap, error) {
	var vmap virtmap.Vmap
	err := vmap.Load(virsh_file)
	if err != nil {
		return vmap, err
	}
	return vmap, nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	vmap, err := GetNodes()
	if err != nil {
		log.Printf("Problem getting vmap: %s", err.Error())
		http.Error(w, `{"error": "Data source error"}`, http.StatusInternalServerError)
		return
	}
	if len(r.URL.Path) < len(api_prefix) {
		log.Printf("Bad request URL: %s", r.URL.Path)
		http.Error(w, `{"error": "Bad request URL"}`, http.StatusNotFound)
		return
	}
	node := strings.TrimLeft(r.URL.Path[len(api_prefix):], "/")

	var encoded []byte

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

	w.Header().Set("Server", "Virtmapper 0.0.2")
	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

func Serve() {
	http.HandleFunc("/api/v1/", handleRequest)
	log.Println("Started")
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
