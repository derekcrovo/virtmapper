package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/subsonic74/virtmapper/virtmap"
)

const virsh_file = "virsh.txt"

func handleRequest(w http.ResponseWriter, r *http.Request) {
	v, err := ioutil.ReadFile(virsh_file)
	if err != nil {
		log.Fatal("Couldn't read file", virsh_file)
	}
	vmap := virtmap.ParseVirsh(string(v))
	log.Printf("Request for %s, virtmap: %d hosts", r.URL.Path, len(vmap))
	data, err := json.MarshalIndent(&vmap, " ", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Println(string(data))
	w.Header().Set("Server", "Virtmapper 0.0.1")
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func Serve() {
	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	log.Println("Starting")
	Serve()
}
