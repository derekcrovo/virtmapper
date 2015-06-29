package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/subsonic74/virtmapper/virtmap"
)

func TestHandleRequest(t *testing.T) {
	BadGetNodes := func() (virtmap.Vmap, error) {
		return virtmap.Vmap{}, errors.New("GetNodes() error")
	}
	GoodGetNodes := func() (virtmap.Vmap, error) {
		return virtmap.Vmap{
				map[string]virtmap.VHost{
					"kvm09": virtmap.VHost{"up", []string{"olh", "tam"}},
					"kvm43": virtmap.VHost{"up", []string{"compute-64"}},
					"kvm30": virtmap.VHost{"down", []string(nil)},
					"kvm59": virtmap.VHost{"up", []string(nil)},
				},
				map[string]virtmap.VGuest{
					"tam":        virtmap.VGuest{"running", "kvm09"},
					"olh":        virtmap.VGuest{"shut", "kvm09"},
					"compute-64": virtmap.VGuest{"paused", "kvm43"},
				},
			},
			nil
	}

	full := `{"hosts":{"kvm09":{"state":"up","guests":["olh","tam"]},"kvm30":{"state":"down","guests":null},"kvm43":{"state":"up","guests":["compute-64"]},"kvm59":{"state":"up","guests":null}},"guests":{"compute-64":{"state":"paused","host":"kvm43"},"olh":{"state":"shut","host":"kvm09"},"tam":{"state":"running","host":"kvm09"}}}`

	tests := []struct {
		getter func() (virtmap.Vmap, error)
		method string
		req    string
		code   int
		body   string
	}{
		{GoodGetNodes, "GET", "/kvm09", http.StatusNotFound, `{"error":"Bad request URL"}`},
		{GoodGetNodes, "GET", "/api/v1/missingnode", http.StatusOK, `{"error":"Node missingnode not found"}`},
		{GoodGetNodes, "GET", "/api/v1/kvm09", http.StatusOK, `{"hosts":{"kvm09":{"state":"up","guests":["olh","tam"]}},"guests":null}`},
		{GoodGetNodes, "GET", "/api/v1/olh", http.StatusOK, `{"hosts":null,"guests":{"olh":{"state":"shut","host":"kvm09"}}}`},
		{GoodGetNodes, "GET", "/api/v1/", http.StatusOK, full},
		{GoodGetNodes, "GET", "/api/v1", http.StatusOK, full},
		{BadGetNodes, "GET", "/api/v1/olh", http.StatusInternalServerError, `{"error":"Data source error"}`},
	}
	buffer := new(bytes.Buffer)
	for _, test := range tests {
		GetNodes = test.getter
		request, _ := http.NewRequest(test.method, test.req, nil)
		response := httptest.NewRecorder()

		handleRequest(response, request)

		if response.Code != test.code {
			t.Fatalf("Unexpected status code %d. Expected: %d", response.Code, http.StatusOK)
		}
		err := json.Compact(buffer, response.Body.Bytes())
		if err != nil {
			t.Fatalf("JSON Compact() error: %v\n%v\nOn request for: %v\n", err, response.Body, test.req)
		}
		if buffer.String() != test.body {
			t.Fatalf("Incorrect API response\nGot:\n%v\nExpected:\n%v\nOn request for: %s", buffer.String(), test.body, test.req)
		}
		buffer.Reset()
	}
}
