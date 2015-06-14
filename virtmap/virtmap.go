package virtmap

import (
	"errors"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

type Host struct {
	Name  string `json:"name"`
	State string `json:"state"`
	VHost string `json:"kvm"`
}

type ByName []Host

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func ReadVirsh(virshFilename string) ([]byte, error) {
	v, err := ioutil.ReadFile(virshFilename)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func ParseVirsh(virshOutput []byte) []Host {
	hosts := make([]Host, 0)
	hostname := ""
	for _, line := range strings.Split(string(virshOutput), "\n") {
		if strings.Contains(line, " | ") {
			hostname = strings.Split(strings.Fields(line)[0], ".")[0]
			if strings.Contains(line, "Name or service not known") {
				continue
			}
			state := "up"
			if strings.Contains(line, "FAILED: timed out") {
				state = "down"
			}
			hosts = append(hosts, Host{Name: hostname, State: state, VHost: ""})
		}
		if hostname != "" {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			if _, err := strconv.Atoi(fields[0]); err == nil || fields[0] == "-" {
				hosts = append(hosts, Host{Name: fields[1], State: fields[2], VHost: hostname})
			}
		}
	}
	sort.Sort(ByName(hosts))
	return hosts
}

func GetHosts(virshFilename string) ([]Host, error) {
	raw, err := ReadVirsh(virshFilename)
	if err != nil {
		return nil, err
	}
	return ParseVirsh(raw), nil
}

func Get(hosts []Host, target string) (Host, []Host, error) {
	guests := make([]Host, 0)
	var host Host
	for _, h := range hosts {
		if h.Name == target {
			host = h
		}
		if h.VHost == target {
			guests = append(guests, h)
		}
	}
	if host.Name == "" {
		return host, guests, errors.New("Host not found")
	}
	return host, guests, nil
}

func HostFor(hosts []Host, target string) (string, error) {
	for _, h := range hosts {
		if h.Name == target && h.VHost != "" {
			return h.VHost, nil
		}
	}
	return "", errors.New("Host not found in slice")
}

func Info(hosts []Host, target string) string {
	host, guests, err := Get(hosts, target)
	if err != nil {
		return "Host " + target + " not found"
	}
	var info string
	if len(guests) != 0 {
		info = host.Name + " is a virtual host for guests:"
		sort.Sort(ByName(guests))
		for i, g := range guests {
			info += " " + g.Name
			if i != len(guests)-1 {
				info += ","
			}
		}
	} else {
		info = target + " is a virtual guest on host: " + host.VHost
	}
	return info
}
