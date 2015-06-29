package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/subsonic74/virtmapper/virtmap"
)

var tests = map[string]struct {
	Resp  *http.Response
	Vmap  virtmap.Vmap
	Error error
}{
	"http://TESTHOST/api/v1/kvm09": {
		&http.Response{
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
		virtmap.Vmap{
			map[string]virtmap.VHost{
				"kvm09": virtmap.VHost{"up", []string{"olh", "tam"}},
			},
			map[string]virtmap.VGuest(nil),
		},
		nil,
	},

	"http://TESTHOST/api/v1/tam": {
		&http.Response{
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
		virtmap.Vmap{
			map[string]virtmap.VHost(nil),
			map[string]virtmap.VGuest{
				"tam": virtmap.VGuest{"running", "kvm09"},
			},
		},
		nil,
	},

	"http://TESTHOST/api/v1/": {
		&http.Response{
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
		virtmap.Vmap{
			map[string]virtmap.VHost{
				"kvm09": virtmap.VHost{"up", []string{"olh", "tam"}},
			},
			map[string]virtmap.VGuest{
				"olh": virtmap.VGuest{"running", "kvm09"},
				"tam": virtmap.VGuest{"paused", "kvm09"},
			},
		},
		nil,
	},
}

func TestQuery(t *testing.T) {
	httpServer = "TESTHOST"
	apiPath := "http://TESTHOST/api/v1/"

	HTTPGetter = func(url string) (*http.Response, error) {
		test, found := tests[url]
		if !found {
			t.Fatalf("Bad Query() test: %#v", url)
		}
		return test.Resp, test.Error
	}

	for req, test := range tests {
		req = req[len(apiPath):]
		resp, err := Query(req)
		if err != test.Error {
			t.Fatalf("Query() returned the wrong error\nGot:\n%v\nExpected:\n%v", err, test.Error)
		}
		if !reflect.DeepEqual(resp, test.Vmap) {
			t.Fatalf("Query() returned bad Vmap\nGot:\n%#v\nExpected:\n%#v", resp, test.Vmap)
		}
	}
}
