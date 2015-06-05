package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/subsonic74/virtmapper/virtmap"
)

const virsh_file = "virsh.txt"

func handleRequest (w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hey %s!", r.URL.Path)
}

func Serve() {
	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	log.Printf("Starting")
//	restserv.Serve()

	v, err := ioutil.ReadFile(virsh_file)
	if err != nil {
		log.Fatal("Couldn't read file", virsh_file)
	}
	vmap := virtmap.ParseVirsh(string(v))
	log.Println(virtmap.Info(vmap, "vhost40"))
	log.Println("info", virtmap.Info(vmap, "yb-mb1"))
}