package virtmap

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
	if len(vmap) == 0 {
		t.Fatal("ParseVirsh() returned nothing")
	}
	expected := []Node{
		{"compute-64", "paused", "kvm43"},
		{"kvm09", "up", ""},
		{"kvm30", "down", ""},
		{"kvm43", "up", ""},
		{"kvm59", "up", ""},
		{"olh", "shut", "kvm09"},
		{"tam", "running", "kvm09"},
	}
	if !reflect.DeepEqual(vmap, Vmap(expected)) {
		t.Fatalf("ParseVirsh() failed.\nGot:\n%v\nExpected:\n%v", vmap, expected)
	}
}

func TestGet(t *testing.T) {
	var vmap Vmap
	vmap.ParseVirsh(virshOutput)
	node, guests, err := vmap.Get("kvm43")
	expected := Node{"kvm43", "up", ""}
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}
	if !reflect.DeepEqual(node, expected) {
		t.Fatalf("Get() returned bad node data\nGot:\n%v\nExpected:\n%v", vmap, expected)
	}
	expectedSlice := []Node{{"compute-64", "paused", "kvm43"}}
	if !reflect.DeepEqual(guests, expectedSlice) {
		t.Fatalf("Get() didn't return the correct guests\nGot:\n%v\nExpected:\n%v", guests, expectedSlice)
	}
	node, guests, err = vmap.Get("olh")
	expected = Node{"olh", "shut", "kvm09"}
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}
	if !reflect.DeepEqual(node, expected) {
		t.Fatalf("Get() returned bad info for test guest\nGot:\n%v\nExpected:\n%v", vmap, expected)
	}
	if len(guests) != 0 {
		t.Fatal("Get() returned guests for a guest")
	}
	node, guests, err = vmap.Get("nonsuch")
	if err == nil {
		t.Fatalf("Get() didn't return an error for a missing node: %v", err)
	}
	if !reflect.DeepEqual(node, Node{"", "", ""}) {
		t.Fatal("Get() returned some node data for a missing node")
	}
	if len(guests) != 0 {
		t.Fatal("Get() returned guests for a missing node")
	}
	node, guests, err = vmap.Get("kvm59")
	if err != nil {
		t.Fatalf("Get() returned an error for an empty host: %v", err)
	}
	if !reflect.DeepEqual(node, Node{"kvm59", "up", ""}) {
		t.Fatal("Get() returned bad data for a missing node")
	}
	if len(guests) != 0 {
		t.Fatal("Get() returned guests for an empty host")
	}
}

func TestVhostFor(t *testing.T) {
	var vmap Vmap
	vmap.ParseVirsh(virshOutput)
	mynode, err := vmap.VhostFor("missing")
	if err == nil {
		t.Fatal("VhostFor() didn't error on missing node")
	}
	mynode, err = vmap.VhostFor("compute-64")
	if err != nil {
		t.Fatal("VhostFor() didn't find node compute-64")
	}
	mynode, err = vmap.VhostFor("kvm43")
	if err == nil {
		t.Fatalf("Found vhost %s for node which isn't virtual", mynode)
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
		{"kvm59", "kvm59 is a virtual host for guests:"},
	}
	for _, test := range tests {
		returned := vmap.Info(test.node)
		if returned == "" {
			t.Fatalf("Info returned nothing for node %s\n", test.node)
		}
		if returned != test.info {
			t.Fatalf("Info() problem\nGot:\n%v\nExpected:\n%v", returned, test.info)
		}
	}
}
