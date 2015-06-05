package virtmap

import (
	"testing"

	"github.com/aryann/difflib"
)

const virt_output = `vhost09.corp.airwave.com | success | rc=0 >>
 Id    Name                           State
----------------------------------------------------
 4     nickel                         running
 -     copper                         shut off
 
 vhost43.corp.airwave.com | success | rc=0 >>
 Id    Name                           State
----------------------------------------------------
 99    amp-integration-64             running
 100   amp-integration-62             running
 101   amp-integration-65             running
 102   amp-integration-63             running
 128   fr-mb1                         running
 129   fr-mb2                         running
 130   fr-mb3                         running
 `

func TestParseVirsh(t *testing.T) {
	hosts := ParseVirsh(virt_output)
	if len(hosts) == 0 {
		t.Error("ParseVirsh returned nothing")
	}
	guests, exists := hosts["vhost09"]
	if !exists {
		t.Error("ParseVirsh didn't find the test host")
	}
	if guests[0].name != "nickel" || guests[0].state != "running" {
		t.Error("ParseVirsh returned bad data")
	}
	if _, exists := hosts["notme"]; exists {
		t.Error("ParseVirsh found non-existent host")
	}
}

func TestHostFor(t *testing.T) {
	hosts := ParseVirsh(virt_output)
	myhost, err := HostFor(hosts, "fr-mb2")
	if err != nil {
		t.Error("HostFor didn't find the test host")
	}
	if myhost != "vhost43" {
		t.Error("HostFor didn't find the test host")
	}
}

func TestInfo(t *testing.T) {
	hosts := ParseVirsh(virt_output)
	info := Info(hosts, "vhost09")
	if info == "" {
		t.Error("Info returned nothing")
	}
	expected := "vhost09 is a virtual host for guests: copper, nickel"
	if info != expected {
		t.Error("Info didn't return the expected string:")
		for _, d := range difflib.Diff([]string{expected}, []string{info}) {
			t.Error(d)
		}
	}
}
