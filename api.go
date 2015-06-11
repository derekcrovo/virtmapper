package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/subsonic74/virtmapper/virtmap"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	v, err := ioutil.ReadFile(virsh_file)
	if err != nil {
		log.Println("Couldn't read file", virsh_file)
		http.Error(w, "Data source problem", 500)
		return
	}
	vmap := virtmap.ParseVirsh(string(v))
	host := r.URL.Path[len("/api/v1/"):]
	var response []byte
	if host == "" {
		log.Printf("Request for entire map, virtmap: %d hosts", len(vmap))
		response, err = json.MarshalIndent(&vmap, " ", "  ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		log.Printf("Request for %s, virtmap: %d hosts", host, len(vmap))
		single := make(map[string][]virtmap.Guest)
		single[host] = vmap[host]
		response, err = json.MarshalIndent(single, " ", "  ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	w.Header().Set("Server", "Virtmapper 0.0.1")
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func Serve() {
	http.HandleFunc("/api/v1/", handleRequest)
	log.Println("Started")
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
