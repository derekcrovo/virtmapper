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
	vmap := &Vmap{
		map[string]VHost{
			"kvm09": VHost{"up", []string{"olh", "tam"}},
			"kvm43": VHost{"up", []string{"compute-64"}},
			"kvm30": VHost{"down", []string(nil)},
			"kvm59": VHost{"up", []string(nil)},
		},
		map[string]VGuest{
			"tam":        VGuest{"running", "kvm09"},
			"olh":        VGuest{"shut", "kvm09"},
			"compute-64": VGuest{"paused", "kvm43"},
		},
	}
	full := `{"hosts":{"kvm09":{"state":"up","guests":["olh","tam"]},"kvm30":{"state":"down","guests":null},"kvm43":{"state":"up","guests":["compute-64"]},"kvm59":{"state":"up","guests":null}},"guests":{"compute-64":{"state":"paused","host":"kvm43"},"olh":{"state":"shut","host":"kvm09"},"tam":{"state":"running","host":"kvm09"}}}`

	tests := []struct {
		testmap *Vmap
		method  string
		req     string
		code    int
		body    string
	}{
		{nil, "GET", "/kvm09", http.StatusNotFound, `{"error":"Bad request URL: /kvm09"}`},
		{vmap, "GET", "/api/v1/missingnode", http.StatusNotFound, `{"error":"Node missingnode not found"}`},
		{vmap, "GET", "/api/v1/kvm09", http.StatusOK, `{"hosts":{"kvm09":{"state":"up","guests":["olh","tam"]}},"guests":null}`},
		{vmap, "GET", "/api/v1/olh", http.StatusOK, `{"hosts":null,"guests":{"olh":{"state":"shut","host":"kvm09"}}}`},
		{vmap, "GET", "/api/v1/", http.StatusOK, full},
	}
	buffer := new(bytes.Buffer)
	mapCh := make(chan *Vmap, 1)
	v := server{mapCh}
	for _, tt := range tests {
		request, _ := http.NewRequest(tt.method, tt.req, nil)
		response := httptest.NewRecorder()

		if tt.testmap != nil {
			mapCh <- tt.testmap
		}

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
	}
}
