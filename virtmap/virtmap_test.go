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
	hosts := ParseVirsh(virshOutput)
	if len(hosts) == 0 {
		t.Fatal("ParseVirsh() returned nothing")
	}
	expected := []Host{
		{"compute-64", "paused", "kvm43"},
		{"kvm09", "up", ""},
		{"kvm30", "down", ""},
		{"kvm43", "up", ""},
		{"olh", "shut", "kvm09"},
		{"tam", "running", "kvm09"},
	}
	if !reflect.DeepEqual(hosts, expected) {
		t.Fatalf("ParseVirsh() failed.\nGot:\n%v\nExpected:\n%v", hosts, expected)
	}
}

func TestGet(t *testing.T) {
	hosts := ParseVirsh(virshOutput)
	host, guests, err := Get(hosts, "kvm43")
	expected := Host{"kvm43", "up", ""}
	if err != nil {
		t.Fatalf("Get() returned an error: %s", err.Error())
	}
	if !reflect.DeepEqual(host, expected) {
		t.Fatalf("Get() returned bad host data\nGot:\n%v\nExpected:\n%v", host, expected)
	}
	expectedSlice := []Host{{"compute-64", "paused", "kvm43"}}
	if !reflect.DeepEqual(guests, expectedSlice) {
		t.Fatalf("Get() didn't return the correct guests\nGot:\n%v\nExpected:\n%v", guests, expectedSlice)
	}
	host, guests, err = Get(hosts, "olh")
	expected = Host{"olh", "shut", "kvm09"}
	if err != nil {
		t.Fatalf("Get() returned an error: %s", err.Error())
	}
	if !reflect.DeepEqual(host, expected) {
		t.Fatalf("Get() returned bad info for test guest\nGot:\n%v\nExpected:\n%v", host, expected)
	}
	if len(guests) != 0 {
		t.Fatal("Get() returned guests for a guest")
	}
	host, guests, err = Get(hosts, "nonsuch")
	if err == nil {
		t.Fatal("Get() didn't return an error for a missing host")
	}
	if !reflect.DeepEqual(host, Host{"", "", ""}) {
		t.Fatal("Get() returned some host data for a missing host")
	}
	if len(guests) != 0 {
		t.Fatal("Get() returned guests for a missing host")
	}
}

func TestHostFor(t *testing.T) {
	hosts := ParseVirsh(virshOutput)
	myhost, err := HostFor(hosts, "missing")
	if err == nil {
		t.Fatal("HostFor() didn't error on missing host")
	}
	myhost, err = HostFor(hosts, "compute-64")
	if err != nil {
		t.Fatal("HostFor() didn't find host compute-64")
	}
	if myhost != "kvm43" {
		t.Fatal("HostFor() didn't find the test host")
	}
	myhost, err = HostFor(hosts, "kvm43")
	if err == nil {
		t.Fatalf("Found host %s for host which isn't virtual", myhost)
	}
}

func TestInfo(t *testing.T) {
	hosts := ParseVirsh(virshOutput)
	info := Info(hosts, "kvm09")
	if info == "" {
		t.Fatal("Info returned nothing")
	}
	expected := "kvm09 is a virtual host for guests: olh, tam"
	if info != expected {
		t.Fatalf("Info() problem\nGot:\n%v\nExpected:\n%v", info, expected)
	}
	info = Info(hosts, "tam")
	if info == "" {
		t.Fatalf("Info() returned nothing")
	}
	expected = "tam is a virtual guest on host: kvm09"
	if info != expected {
		t.Fatalf("Info() problem\nGot:\n%v\nExpected:\n%v", info, expected)
	}
	info = Info(hosts, "gone")
	if info == "" {
		t.Fatal("Info returned nothing")
	}
	expected = "Host gone not found"
	if info != expected {
		t.Fatalf("Info() problem\nGot:\n%v\nExpected:\n%v", info, expected)
	}
}
