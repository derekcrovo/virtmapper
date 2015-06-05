package virtmap

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

type Guest struct {
	name string
	state string
}

type ByName []Guest

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].name < a[j].name }

func ParseVirsh(virshOutput string) map[string][]Guest {
	guestMap := make(map[string][]Guest)
	host := ""
	for _, l := range(strings.Split(virshOutput, "\n")) {
		if strings.Contains(l, "success") {
			host = strings.Split(strings.Fields(l)[0], ".")[0]
		} else if l == "" {
			host = ""
		}
		if host != "" {
			fields := strings.Fields(l)
			if len(fields) == 0 {
				continue
			}
			if _, err := strconv.Atoi(fields[0]); err == nil || fields[0] == "-" {
				guestMap[host] = append(guestMap[host],Guest{fields[1], fields[2]})
			}
		}
	}
	return guestMap
}

func HostFor(hosts map[string][]Guest, guest string) (string, error) {
	for h, gs := range(hosts) {
		for _, g := range(gs) {
			if g.name == guest {
				return h, nil
			}
		}
	}
	return "", errors.New("Guest not found in map")
}

func Info(hosts map[string][]Guest, system string) string {
	if guests, exists := hosts[system]; exists {
		info := system + " is a virtual host for guests:"
		sort.Sort(ByName(guests))
		for i, g := range(guests) {
			info += " " + g.name
			if i != len(guests) - 1 {
				info += ","
			}
		}
		return info
	} else {
		info := system + " is a virtual guest on host: "
		host, err := HostFor(hosts, system); if  err == nil {
			info += host
			return info
		}
	}
	return system + " is not known"
}