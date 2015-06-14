package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/subsonic74/virtmapper/virtmap"
)

func TestHandleRequest(t *testing.T) {
	GetHosts = func() ([]virtmap.Host, error) {
		return []virtmap.Host{
			{"compute-64", "paused", "kvm43"},
			{"kvm09", "up", ""},
			{"kvm30", "down", ""},
			{"kvm43", "up", ""},
			{"olh", "shut", "kvm09"},
			{"tam", "running", "kvm09"},
		}, nil
	}

	tests := []struct {
		method string
		req    string
		code   int
		body   string
	}{
		{"GET", "/api/v1/missinghost", http.StatusOK, "{'error': 'Host not found'}"},
		{"GET", "/api/v1/kvm09", http.StatusOK, `{
   "host": {
     "name": "kvm09",
     "state": "up",
     "kvm": ""
   },
   "guests": [
     {
       "name": "olh",
       "state": "shut",
       "kvm": "kvm09"
     },
     {
       "name": "tam",
       "state": "running",
       "kvm": "kvm09"
     }
   ]
 }`},
		{"GET", "/api/v1/olh", http.StatusOK, `{
   "host": {
     "name": "olh",
     "state": "shut",
     "kvm": "kvm09"
   },
   "guests": []
 }`},
	}
	for _, test := range tests {
		request, _ := http.NewRequest(test.method, test.req, nil)
		response := httptest.NewRecorder()

		handleRequest(response, request)

		if response.Code != test.code {
			t.Fatalf("Non-expected status code: %v\n\tbody: %v", http.StatusOK, response.Code)
		}
		if response.Body.String() != test.body {
			t.Fatalf("Incorrect API response\n%v\nexpected\n%v", response.Body, test.body)
		}
	}
}
