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
		return []virtmap.Node{{}}, errors.New("GetNodes() error")
	}
	GoodGetNodes := func() (virtmap.Vmap, error) {
		return virtmap.Vmap{
			{"compute-64", "paused", "kvm43"},
			{"kvm09", "up", ""},
			{"kvm30", "down", ""},
			{"kvm43", "up", ""},
			{"olh", "shut", "kvm09"},
			{"tam", "running", "kvm09"},
		}, nil
	}

	full := `{"vmap":[{"name":"compute-64","state":"paused","vhost":"kvm43"},{"name":"kvm09","state":"up","vhost":""},{"name":"kvm30","state":"down","vhost":""},{"name":"kvm43","state":"up","vhost":""},{"name":"olh","state":"shut","vhost":"kvm09"},{"name":"tam","state":"running","vhost":"kvm09"}]}`

	tests := []struct {
		getter func() (virtmap.Vmap, error)
		method string
		req    string
		code   int
		body   string
	}{
		{GoodGetNodes, "GET", "/kvm09", http.StatusNotFound, `{"error":"Bad request URL"}`},
		{GoodGetNodes, "GET", "/api/v1/missingnode", http.StatusOK, `{"error":"Node missingnode not found"}`},
		{GoodGetNodes, "GET", "/api/v1/kvm09", http.StatusOK, `{"node":{"name":"kvm09","state":"up","vhost":""},"guests":[{"name":"olh","state":"shut","vhost":"kvm09"},{"name":"tam","state":"running","vhost":"kvm09"}]}`},
		{GoodGetNodes, "GET", "/api/v1/olh", http.StatusOK, `{"node":{"name":"olh","state":"shut","vhost":"kvm09"},"guests":[]}`},
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
			t.Fatalf("JSON Compact() error: %v\n%v", err, response.Body)
		}
		if buffer.String() != test.body {
			t.Fatalf("Incorrect API response\nGot:\n%v\nExpected:\n%v", buffer.String(), test.body)
		}
		buffer.Reset()
	}
}
