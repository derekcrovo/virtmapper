package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

var tests = map[string]struct {
	Resp  *http.Response
	Vmap  Vmap
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
		Vmap{
			map[string]VHost{
				"kvm09": VHost{"up", []string{"olh", "tam"}},
			},
			map[string]VGuest(nil),
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
		Vmap{
			map[string]VHost(nil),
			map[string]VGuest{
				"tam": VGuest{"running", "kvm09"},
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
		Vmap{
			map[string]VHost{
				"kvm09": VHost{"up", []string{"olh", "tam"}},
			},
			map[string]VGuest{
				"olh": VGuest{"running", "kvm09"},
				"tam": VGuest{"paused", "kvm09"},
			},
		},
		nil,
	},
}

func TestQuery(t *testing.T) {
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
		resp, err := Query("TESTHOST", req)
		if err != test.Error {
			t.Fatalf("Query() returned the wrong error\nGot:\n%v\nExpected:\n%v", err, test.Error)
		}
		if !reflect.DeepEqual(resp, test.Vmap) {
			t.Fatalf("Query() returned bad Vmap\nGot:\n%#v\nExpected:\n%#v", resp, test.Vmap)
		}
	}
}
