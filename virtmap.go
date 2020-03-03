package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

// errNodeNotFound is returned when the requested host is not present in the vmap
var errNodeNotFound = errors.New("Node not found")

// VHost is a virtual host which contains several virtual guests
// State may be "up" or "down"
type VHost struct {
	State  string   `json:"state"`
	Guests []string `json:"guests"`
}

// VGuest is a virtual guest. Includes the name of its virtual host
// State may be "running", "paused", or "shut" (i.e. shut down) as reported by "virsh list"
type VGuest struct {
	State string `json:"state"`
	Host  string `json:"host"`
}

// Vmap is the main virtual map.  It contains a map of guests and
// a map of hosts to support queries in either direction.
type Vmap struct {
	Hosts  map[string]VHost  `json:"hosts"`
	Guests map[string]VGuest `json:"guests"`
}

// Length returns the total number of hosts in the map
func (v Vmap) Length() int {
	return len(v.Hosts) + len(v.Guests)
}

// ParseVirsh parses the output of an Ansible run of
// "virsh list --all" on all the virtual hosts
func (v *Vmap) ParseVirsh(virshOutput []byte) {
	v.Hosts = make(map[string]VHost)
	v.Guests = make(map[string]VGuest)
	nodename := ""
	for _, line := range strings.Split(string(virshOutput), "\n") {
		// Ansible status lines contain the hostname and any connection errors
		if strings.Contains(line, " | ") {
			nodename = strings.Split(strings.Fields(line)[0], ".")[0]
			if strings.Contains(line, "Name or service not known") {
				continue
			}
			state := "up"
			if strings.Contains(line, "FAILED: timed out") {
				state = "down"
			}
			v.Hosts[nodename] = VHost{State: state}
		}
		if nodename != "" {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			// Guest state lines
			if _, err := strconv.Atoi(fields[0]); err == nil || fields[0] == "-" {
				host := v.Hosts[nodename]
				host.Guests = append(host.Guests, fields[1])
				v.Hosts[nodename] = host
				v.Guests[fields[1]] = VGuest{State: fields[2], Host: nodename}
			}
		}
	}
	for _, h := range v.Hosts {
		sort.Strings(h.Guests)
	}
}

// Load Loads the Ansible output file and parses it into the Vmap
func (v *Vmap) Load(virshFilename string) error {
	raw, err := ioutil.ReadFile(virshFilename)
	if err != nil {
		return err
	}
	v.ParseVirsh(raw)
	return nil
}

// Get returns a host from the map.  The target host
// may be a virtual host or a virtual guest.
// nodeNotFoundErr is returned when the target is not in the map
func (v Vmap) Get(target string) (Vmap, error) {
	var result Vmap
	found := false
	for n, h := range v.Hosts {
		if n == target {
			result.Hosts = make(map[string]VHost)
			result.Hosts[n] = h
			found = true
			break
		}
	}
	for n, g := range v.Guests {
		if n == target {
			result.Guests = make(map[string]VGuest)
			result.Guests[n] = g
			found = true
			break
		}
	}
	if !found {
		return result, errNodeNotFound
	}
	return result, nil
}

// Info returns a friendly text string describing the target host.
// Used in user cli queries.
func (v Vmap) Info(target string) string {
	result, err := v.Get(target)
	if err != nil {
		return fmt.Sprintf("Node %s not found", target)
	}
	var info string
	if h, isHost := result.Hosts[target]; isHost {
		info = fmt.Sprintf("%s is a virtual host for guests: %s", target, strings.Join(h.Guests, ", "))
	}
	if g, isGuest := result.Guests[target]; isGuest {
		info = fmt.Sprintf("%s is a virtual guest on host: %s", target, g.Host)
	}
	return info
}
