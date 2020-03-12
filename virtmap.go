package main

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// VHost is a virtual host which contains several virtual guests
// State may be "up" or "down"
type VHost struct {
	State  string   `json:"state"`
	Guests []string `json:"guests"`
}

// VGuest is a virtual guest. Includes the name of its virtual host
// State may be "running", "paused", or "shut" (i.e. shut down)
// as reported by "virsh list --all"
type VGuest struct {
	State string `json:"state"`
	Host  string `json:"host"`
}

// Vmap is the main virtual map type.  It contains a map of guests
// and a map of hosts to support queries in either direction.
type Vmap struct {
	Hosts  map[string]VHost  `json:"hosts"`
	Guests map[string]VGuest `json:"guests"`
}

// Length returns the total number of hosts in the map
func (v Vmap) Length() int {
	return len(v.Hosts) + len(v.Guests)
}

// Load Loads the Ansible output file and parses it into the Vmap
func (v *Vmap) Load(ansibleOutputFilename string) error {
	raw, err := ioutil.ReadFile(ansibleOutputFilename)
	if err != nil {
		return err
	}
	x := ParseAnsibleOutput(raw)
	v.Hosts, v.Guests = x.Hosts, x.Guests
	return nil
}

// Get returns a host from the map.  The target host
// may be a virtual host or a virtual guest.
// nodeNotFoundErr is returned when the target is not in the map
func (v Vmap) Get(target string) (*Vmap, error) {
	for n, h := range v.Hosts {
		if n == target {
			return &Vmap{Hosts: map[string]VHost{n: h}}, nil
		}
	}
	for n, g := range v.Guests {
		if n == target {
			return &Vmap{Guests: map[string]VGuest{n: g}}, nil
		}
	}
	return nil, ErrNodeNotFound
}

// Info returns a friendly text string describing the target host.
// Used in user cli queries.
func (v *Vmap) Info(target string) string {
	result, err := v.Get(target)
	if err != nil {
		return fmt.Sprintf("Node %s not found", target)
	}
	var info string
	if h, ok := result.Hosts[target]; ok {
		info = fmt.Sprintf("%s is a virtual host for guests: %s", target, strings.Join(h.Guests, ", "))
	}
	if g, ok := result.Guests[target]; ok {
		info = fmt.Sprintf("%s is a virtual guest on host: %s", target, g.Host)
	}
	return info
}

// SafeVmap is a Vmap wrapped with a mutex for the
// server to use since Go maps are not thread safe
type SafeVmap struct {
	sync.RWMutex
	Vmap
}

// Length for SafeVmap wraps Vmap.Length() in a read lock
// s is a pointer receiver so we don't copy the mutex
func (s *SafeVmap) Length() int {
	s.RLock()
	defer s.RUnlock()
	return s.Vmap.Length()
}

// Load for SafeVmap wraps Vmap.Load() in a (write) lock
func (s *SafeVmap) Load(ansibleOutputFilename string) error {
	s.Lock()
	defer s.Unlock()
	return s.Vmap.Load(ansibleOutputFilename)
}

// Get for SafeVmap wraps Vmap.Get() in a read lock
// s is a pointer receiver so we don't copy the mutex
func (s *SafeVmap) Get(target string) (*Vmap, error) {
	s.RLock()
	defer s.RUnlock()
	return s.Vmap.Get(target)
}

// Info for SafeVmap wraps Vmap.Info() in a read lock
// s is a pointer receiver so we don't copy the mutex
func (s *SafeVmap) Info(target string) string {
	s.RLock()
	defer s.RUnlock()
	return s.Vmap.Info(target)
}

// ParseAnsibleOutput parses the output of an Ansible run of
// "virsh list --all" on all the virtual hosts
func ParseAnsibleOutput(ansibleOutput []byte) *Vmap {
	v := &Vmap{
		Hosts:  make(map[string]VHost),
		Guests: make(map[string]VGuest),
	}
	nodename := ""
	for _, line := range strings.Split(string(ansibleOutput), "\n") {
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
	return v
}
