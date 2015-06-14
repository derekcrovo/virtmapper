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
`)

func TestParseVirsh(t *testing.T) {
	nodes := ParseVirsh(virshOutput)
	if len(nodes) == 0 {
		t.Fatal("ParseVirsh() returned nothing")
	}
	expected := []Node{
		{"compute-64", "paused", "kvm43"},
		{"kvm09", "up", ""},
		{"kvm30", "down", ""},
		{"kvm43", "up", ""},
		{"olh", "shut", "kvm09"},
		{"tam", "running", "kvm09"},
	}
	if !reflect.DeepEqual(nodes, expected) {
		t.Fatalf("ParseVirsh() failed.\nGot:\n%v\nExpected:\n%v", nodes, expected)
	}
}

func TestGet(t *testing.T) {
	nodes := ParseVirsh(virshOutput)
	node, guests, err := Get(nodes, "kvm43")
	expected := Node{"kvm43", "up", ""}
	if err != nil {
		t.Fatalf("Get() returned an error: %s", err.Error())
	}
	if !reflect.DeepEqual(node, expected) {
		t.Fatalf("Get() returned bad node data\nGot:\n%v\nExpected:\n%v", node, expected)
	}
	expectedSlice := []Node{{"compute-64", "paused", "kvm43"}}
	if !reflect.DeepEqual(guests, expectedSlice) {
		t.Fatalf("Get() didn't return the correct guests\nGot:\n%v\nExpected:\n%v", guests, expectedSlice)
	}
	node, guests, err = Get(nodes, "olh")
	expected = Node{"olh", "shut", "kvm09"}
	if err != nil {
		t.Fatalf("Get() returned an error: %s", err.Error())
	}
	if !reflect.DeepEqual(node, expected) {
		t.Fatalf("Get() returned bad info for test guest\nGot:\n%v\nExpected:\n%v", node, expected)
	}
	if len(guests) != 0 {
		t.Fatal("Get() returned guests for a guest")
	}
	node, guests, err = Get(nodes, "nonsuch")
	if err == nil {
		t.Fatal("Get() didn't return an error for a missing node")
	}
	if !reflect.DeepEqual(node, Node{"", "", ""}) {
		t.Fatal("Get() returned some node data for a missing node")
	}
	if len(guests) != 0 {
		t.Fatal("Get() returned guests for a missing node")
	}
}

func TestNodeFor(t *testing.T) {
	nodes := ParseVirsh(virshOutput)
	mynode, err := NodeFor(nodes, "missing")
	if err == nil {
		t.Fatal("NodeFor() didn't error on missing node")
	}
	mynode, err = NodeFor(nodes, "compute-64")
	if err != nil {
		t.Fatal("NodeFor() didn't find node compute-64")
	}
	if mynode != "kvm43" {
		t.Fatal("NodeFor() didn't find the test node")
	}
	mynode, err = NodeFor(nodes, "kvm43")
	if err == nil {
		t.Fatalf("Found node %s for node which isn't virtual", mynode)
	}
}

func TestInfo(t *testing.T) {
	nodes := ParseVirsh(virshOutput)
	info := Info(nodes, "kvm09")
	if info == "" {
		t.Fatal("Info returned nothing")
	}
	expected := "kvm09 is a virtual node for guests: olh, tam"
	if info != expected {
		t.Fatalf("Info() problem\nGot:\n%v\nExpected:\n%v", info, expected)
	}
	info = Info(nodes, "tam")
	if info == "" {
		t.Fatalf("Info() returned nothing")
	}
	expected = "tam is a virtual guest on node: kvm09"
	if info != expected {
		t.Fatalf("Info() problem\nGot:\n%v\nExpected:\n%v", info, expected)
	}
	info = Info(nodes, "gone")
	if info == "" {
		t.Fatal("Info returned nothing")
	}
	expected = "Node gone not found"
	if info != expected {
		t.Fatalf("Info() problem\nGot:\n%v\nExpected:\n%v", info, expected)
	}
}
