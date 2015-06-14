package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/subsonic74/virtmapper/virtmap"
)

const api_prefix = "/api/v1"

type apiHostResponse struct {
	Host   virtmap.Host   `json:"host"`
	Guests []virtmap.Host `json:"guests"`
}

type apiFullResponse struct {
	Hosts []virtmap.Host `json:"hosts"`
}

var GetHosts = func() ([]virtmap.Host, error) { return virtmap.GetHosts(virsh_file) }

func handleRequest(w http.ResponseWriter, r *http.Request) {
	hosts, err := GetHosts()
	if err != nil {
		log.Printf("Problem getting hosts: %s", err.Error())
		http.Error(w, `{"error": "Data source error"}`, http.StatusInternalServerError)
		return
	}
	if len(r.URL.Path) < len(api_prefix) {
		log.Printf("Bad request URL: %s", r.URL.Path)
		http.Error(w, `{"error": "Bad request URL"}`, http.StatusNotFound)
		return
	}
	host := strings.TrimLeft(r.URL.Path[len(api_prefix):], "/")
	var encoded []byte
	if host == "" {
		log.Printf("Request for entire map, virtmap: %d hosts", len(hosts))
		var response apiFullResponse
		response.Hosts = hosts
		encoded, err = json.MarshalIndent(response, " ", "  ")
		if err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}"`, http.StatusInternalServerError)
			return
		}
	} else {
		log.Printf("Request for %s, virtmap: %d hosts", host, len(hosts))
		var response apiHostResponse
		response.Host, response.Guests, err = virtmap.Get(hosts, host)
		if err != nil {
			encoded = []byte(`{"error": "Host ` + host + ` not found"}`)
		} else {
			encoded, err = json.MarshalIndent(response, " ", "  ")
			if err != nil {
				http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
				return
			}
		}
	}
	w.Header().Set("Server", "Virtmapper 0.0.1")
	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

func Serve() {
	http.HandleFunc("/api/v1/", handleRequest)
	log.Println("Started")
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
