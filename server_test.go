package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRequest(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	vmap := Vmap{
		Hosts: map[string]VHost{
			"kvm09": VHost{"up", []string{"olh", "tam"}},
			"kvm43": VHost{"up", []string{"compute-64"}},
			"kvm30": VHost{"down", []string(nil)},
			"kvm59": VHost{"up", []string(nil)},
		},
		Guests: map[string]VGuest{
			"tam":        VGuest{"running", "kvm09"},
			"olh":        VGuest{"shut", "kvm09"},
			"compute-64": VGuest{"paused", "kvm43"},
		},
	}
	full := `{"hosts":{"kvm09":{"state":"up","guests":["olh","tam"]},"kvm30":{"state":"down","guests":null},"kvm43":{"state":"up","guests":["compute-64"]},"kvm59":{"state":"up","guests":null}},"guests":{"compute-64":{"state":"paused","host":"kvm43"},"olh":{"state":"shut","host":"kvm09"},"tam":{"state":"running","host":"kvm09"}}}`

	tests := []struct {
		method string
		req    string
		code   int
		body   string
	}{
		{"GET", "/kvm09", http.StatusNotFound, `{"error":"Bad request URL: /kvm09"}`},
		{"GET", "/api/v1/vmap/missingnode", http.StatusNotFound, `{"error":"Node missingnode not found"}`},
		{"GET", "/api/v1/vmap/kvm09", http.StatusOK, `{"hosts":{"kvm09":{"state":"up","guests":["olh","tam"]}},"guests":null}`},
		{"GET", "/api/v1/vmap/olh", http.StatusOK, `{"hosts":null,"guests":{"olh":{"state":"shut","host":"kvm09"}}}`},
		{"GET", "/api/v1/vmap/", http.StatusOK, full},
		{"POST", "/api/v1/vmap/", http.StatusMethodNotAllowed, `{"error":"Bad request method: POST, only GET is allowed"}`},
	}
	buffer := new(bytes.Buffer)
	v := server{svmap: &SafeVmap{Vmap: vmap}}
	for _, tt := range tests {
		t.Run(tt.req, func(t *testing.T) {
			request, _ := http.NewRequest(tt.method, tt.req, nil)
			response := httptest.NewRecorder()

			v.handleRequest(response, request)

			if response.Code != tt.code {
				t.Fatalf("Unexpected status code %d. Expected: %d for request %s", response.Code, tt.code, tt.req)
			}
			err := json.Compact(buffer, response.Body.Bytes())
			if err != nil {
				t.Fatalf("JSON Compact() error: %v\n%v\nOn request for: %v\n", err, response.Body, tt.req)
			}
			if buffer.String() != tt.body {
				t.Fatalf("Incorrect API response\nGot:\n%v\nExpected:\n%v\nOn request for: %s", buffer.String(), tt.body, tt.req)
			}
			buffer.Reset()
		})
	}
}
