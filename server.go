package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/urfave/cli"
)

const (
	// APIPrefix is the versioned URL endpoint for the server
	APIPrefix = "/api/" + APIVersion + "/"

	// VMAPPrefix is the vmap endpoint URL
	VMAPPrefix = APIPrefix + "vmap/"
)

// ErrNodeNotFound is returned when the requested host is not present in the vmap
var ErrNodeNotFound = errors.New("Node not found")

type server struct {
	svmap *SafeVmap
}

// newServer creates an initialized server struct
func newServer() server {
	return server{
		svmap: &SafeVmap{},
	}
}

// The HTTP handler for the API.  Returns results in JSON format.
func (s *server) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Virtmapper v"+Version)
	if r.Method != "GET" {
		err := fmt.Errorf("Bad request method: %s, only GET is allowed", r.Method)
		log.Println(err)
		s.respondErr(w, r, http.StatusMethodNotAllowed, err)
		return
	}
	if !strings.HasPrefix(r.URL.Path, APIPrefix) {
		err := fmt.Errorf("Bad request URL: %s", r.URL.Path)
		log.Println(err)
		s.respondErr(w, r, http.StatusNotFound, err)
		return
	}
	node := strings.TrimLeft(r.URL.Path[len(VMAPPrefix):], "/")
	if s.svmap.Length() == 0 {
		log.Printf("Vmap is empty")
	}
	var response *Vmap
	if node == "" {
		log.Printf("Request for entire map, virtmap: %d nodes", s.svmap.Length())
		response = &s.svmap.Vmap
	} else {
		log.Printf("Request for %s, virtmap: %d nodes", node, s.svmap.Length())
		var err error
		response, err = s.svmap.Get(node)
		if err == ErrNodeNotFound {
			s.respondErr(w, r, http.StatusNotFound, fmt.Errorf("Node %s not found", node))
			return
		}
		if err != nil {
			s.respondErr(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	s.respond(w, r, http.StatusOK, response)
}

// respond is a helper to respond in JSON
func (s *server) respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("encode response: %s", err)
	}
}

// respondErr is a helper to respond with an error in JSON
func (s *server) respondErr(w http.ResponseWriter, r *http.Request, status int, err error) {
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
func (s *server) Serve(c *cli.Context) {
	done := make(chan struct{})
	s.LaunchReloader(c.String("ansibleOutputFile"), c.Int("refreshInterval"), done)
	http.HandleFunc(APIPrefix, s.handleRequest)
	log.Println("Starting server, listening on", c.String("address"))
	log.Fatal(http.ListenAndServe(c.String("address"), nil))
	close(done)
}

// Reloader launches a goroutine which loads and
// parses the ansibleOutputFile periodically
func (s *server) LaunchReloader(ansibleOutputFile string, refresh int, done chan struct{}) {
	var delay time.Duration
	go func() {
		for {
			select {
			case <-time.After(delay):
				err := s.svmap.Load(ansibleOutputFile)
				if err != nil {
					log.Printf("Problem getting vmap: %s", err.Error())
				}
				log.Printf("Reloaded from %s, %d entries in map.\n", ansibleOutputFile, s.svmap.Length())
				delay = time.Duration(refresh) * time.Minute
			case <-done:
				return
			}
		}
	}()
}
