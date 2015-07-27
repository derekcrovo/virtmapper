package main

import (
	"reflect"
	"testing"
)

var virshOutput = []byte(`kvm21.example.com | FAILED => FAILED: [Errno -2] Name or service not known
kvm09.example.com | success | rc=0 >>
 Id    Name                           State
----------------------------------------------------
 4     tam                            running
 -     olh                            shut off
 
 kvm43.example.com | success | rc=0 >>
 Id    Name                           State
----------------------------------------------------
 99    compute-64                     paused

kvm30.example.com | FAILED => FAILED: timed out
kvm59.example.com | success | rc=0 >>
 Id    Name                           State
----------------------------------------------------

`)

func TestParseVirsh(t *testing.T) {
	var vmap Vmap
	vmap.ParseVirsh(virshOutput)
	if vmap.Length() == 0 {
		t.Fatal("ParseVirsh() returned nothing")
	}
	expected := Vmap{
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
	if !reflect.DeepEqual(vmap, Vmap(expected)) {
		t.Fatalf("ParseVirsh() failed.\nGot:\n%#v\nExpected:\n%#v", vmap, expected)
	}
}

func TestGet(t *testing.T) {
	var vmap Vmap
	vmap.ParseVirsh(virshOutput)
	tests := []struct {
		target string
		result Vmap
		error  string
	}{
		{
			"kvm43",
			Vmap{
				map[string]VHost{"kvm43": VHost{"up", []string{"compute-64"}}},
				map[string]VGuest(nil),
			},
			"",
		},
		{
			"olh",
			Vmap{
				map[string]VHost(nil),
				map[string]VGuest{"olh": VGuest{"shut", "kvm09"}},
			},
			"",
		},
		{
			"nonsuch",
			Vmap{},
			"Node not found",
		},
		{
			"kvm59",
			Vmap{
				map[string]VHost{"kvm59": VHost{"up", []string(nil)}},
				map[string]VGuest(nil),
			},
			"",
		},
	}
	for _, test := range tests {
		node, err := vmap.Get(test.target)
		if test.error != "" {
			if err.Error() != test.error {
				t.Fatalf("Get() returned the wrong error\nGot:\n%v\nExpected:\n%v", err, test.error)
			}
		} else if err != nil {
			t.Fatalf("Get() returned an error unexpectedly: %v", err)
		}
		if !reflect.DeepEqual(node, test.result) {
			t.Fatalf("Get() returned bad node data\nGot:\n%#v\nExpected:\n%#v", node, test.result)
		}
	}
}

func TestInfo(t *testing.T) {
	var vmap Vmap
	vmap.ParseVirsh(virshOutput)
	tests := []struct {
		node string
		info string
	}{
		{"kvm09", "kvm09 is a virtual host for guests: olh, tam"},
		{"tam", "tam is a virtual guest on host: kvm09"},
		{"gone", "Node gone not found"},
		{"kvm59", "kvm59 is a virtual host for guests: "},
	}
	for _, test := range tests {
		returned := vmap.Info(test.node)
		if returned == "" {
			t.Fatalf("Info returned nothing for node %s\n", test.node)
		}
		if returned != test.info {
			t.Fatalf("Info() problem\nGot:\n%#v\nExpected:\n%#v", returned, test.info)
		}
	}
}
