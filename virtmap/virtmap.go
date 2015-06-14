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

type ByName []Node

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

func ParseVirsh(virshOutput []byte) []Node {
	nodes := make([]Node, 0)
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
			nodes = append(nodes, Node{Name: nodename, State: state, VHost: ""})
		}
		if nodename != "" {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			if _, err := strconv.Atoi(fields[0]); err == nil || fields[0] == "-" {
				nodes = append(nodes, Node{Name: fields[1], State: fields[2], VHost: nodename})
			}
		}
	}
	sort.Sort(ByName(nodes))
	return nodes
}

func GetNodes(virshFilename string) ([]Node, error) {
	raw, err := ReadVirsh(virshFilename)
	if err != nil {
		return nil, err
	}
	return ParseVirsh(raw), nil
}

func Get(nodes []Node, target string) (Node, []Node, error) {
	guests := make([]Node, 0)
	var node Node
	for _, h := range nodes {
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

func NodeFor(nodes []Node, target string) (string, error) {
	for _, h := range nodes {
		if h.Name == target && h.VHost != "" {
			return h.VHost, nil
		}
	}
	return "", errors.New("Node not found in slice")
}

func Info(nodes []Node, target string) string {
	node, guests, err := Get(nodes, target)
	if err != nil {
		return "Node " + target + " not found"
	}
	var info string
	if len(guests) != 0 {
		info = node.Name + " is a virtual node for guests:"
		sort.Sort(ByName(guests))
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
