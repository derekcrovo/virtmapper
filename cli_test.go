package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

var tests = map[string]struct {
	Resp  http.Response
	Vmap  Vmap
	Error error
}{
	"http://TESTHOST/api/v1/vmap/kvm09": {
		Resp: http.Response{
			StatusCode: http.StatusOK,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"hosts": {
		"kvm09": {
			"state": "up",
			"guests": [
				"olh",
				"tam"
			]
		}
	},
	"guests": null
}`)),
		},
		Vmap: Vmap{
			Hosts: map[string]VHost{
				"kvm09": VHost{"up", []string{"olh", "tam"}},
			},
			Guests: map[string]VGuest(nil),
		},
		Error: nil,
	},

	"http://TESTHOST/api/v1/vmap/tam": {
		Resp: http.Response{
			StatusCode: http.StatusOK,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"hosts": null,
	"guests": {
		"tam": {
			"state": "running",
			"host": "kvm09"
		}
	}
}`)),
		},
		Vmap: Vmap{
			Hosts: map[string]VHost(nil),
			Guests: map[string]VGuest{
				"tam": VGuest{"running", "kvm09"},
			},
		},
		Error: nil,
	},

	"http://TESTHOST/api/v1/vmap/": {
		Resp: http.Response{
			StatusCode: http.StatusOK,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
{
	"hosts": {
		"kvm09": {
			"state": "up",
			"guests": [
			"olh",
			"tam"
			]
		}
	},
	"guests": {
		"olh": {
		"state": "running",
		"host": "kvm09"
		},
		"tam": {
		"state": "paused",
		"host": "kvm09"
		}
	}
}`)),
		},
		Vmap: Vmap{
			Hosts: map[string]VHost{
				"kvm09": VHost{"up", []string{"olh", "tam"}},
			},
			Guests: map[string]VGuest{
				"olh": VGuest{"running", "kvm09"},
				"tam": VGuest{"paused", "kvm09"},
			},
		},
		Error: nil,
	},
}

func TestQuery(t *testing.T) {
	apiPath := "http://TESTHOST/api/v1/vmap/"

	HTTPGetter = func(url string) (*http.Response, error) {
		test, found := tests[url]
		if !found {
			t.Fatalf("Bad Query() test: URL %q not found", url)
		}
		return &test.Resp, test.Error
	}

	for req, tt := range tests {
		req = req[len(apiPath):]
		t.Run(req, func(t *testing.T) {
			resp, err := Query("TESTHOST", req)
			if err != tt.Error {
				t.Fatalf("Query() returned the wrong error\nGot:\n%v\nExpected:\n%v", err, tt.Error)
			}
			if !reflect.DeepEqual(resp, &tt.Vmap) {
				t.Fatalf("Query() returned bad Vmap\nGot:\n%#v\nExpected:\n%#v", resp, tt.Vmap)
			}
		})
	}
}
