package virtmap

import (
	"errors"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

type Node struct {
	Name  string `json:"name"`
	State string `json:"state"`
	VHost string `json:"vhost"`
}

type Vmap []Node

func (a Vmap) Len() int           { return len(a) }
func (a Vmap) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Vmap) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (v *Vmap) ParseVirsh(virshOutput []byte) {
	nodename := ""
	for _, line := range strings.Split(string(virshOutput), "\n") {
		if strings.Contains(line, " | ") {
			nodename = strings.Split(strings.Fields(line)[0], ".")[0]
			if strings.Contains(line, "Name or service not known") {
				continue
			}
			state := "up"
			if strings.Contains(line, "FAILED: timed out") {
				state = "down"
			}
			*v = append(*v, Node{Name: nodename, State: state, VHost: ""})
		}
		if nodename != "" {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			if _, err := strconv.Atoi(fields[0]); err == nil || fields[0] == "-" {
				*v = append(*v, Node{Name: fields[1], State: fields[2], VHost: nodename})
			}
		}
	}
	sort.Sort(Vmap(*v))
}

func (v *Vmap) Load(virshFilename string) error {
	raw, err := ioutil.ReadFile(virshFilename)
	if err != nil {
		return err
	}
	v.ParseVirsh(raw)
	return nil
}

func (v Vmap) Get(target string) (Node, []Node, error) {
	var node Node
	guests := make([]Node, 0)
	for _, h := range v {
		if h.Name == target {
			node = h
		}
		if h.VHost == target {
			guests = append(guests, h)
		}
	}
	if node.Name == "" {
		return node, guests, errors.New("Node not found")
	}
	return node, guests, nil
}

func (v Vmap) VhostFor(target string) (Node, error) {
	for _, h := range v {
		if h.Name == target && h.VHost != "" {
			return h, nil
		}
	}
	return Node{}, errors.New("Node not found in slice")
}

func (v Vmap) Info(target string) string {
	node, guests, err := v.Get(target)
	if err != nil {
		return "Node " + target + " not found"
	}
	var info string
	if len(guests) != 0 {
		info = node.Name + " is a virtual node for guests:"
		sort.Sort(Vmap(guests))
		for i, g := range guests {
			info += " " + g.Name
			if i != len(guests)-1 {
				info += ","
			}
		}
	} else {
		info = target + " is a virtual guest on node: " + node.VHost
	}
	return info
}
