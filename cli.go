package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/subsonic74/virtmapper/virtmap"
)

func Query(node string) []virtmap.Node {
	response, err := http.Get("http://" + httpServer + "/api/v1/" + query)
	if err != nil {
		fmt.Printf("Get() error, %s\n", err.Error())
		return nil
	}
	var nodes []virtmap.Node
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("ReadAll() error, %s\n", err.Error())
		return nil
	}
	err = json.Unmarshal(body, &nodes)
	if err != nil {
		fmt.Printf("Unmarshal() error, %s\n", err.Error())
		return nil
	}
	return nodes
}

func Display(nodes []virtmap.Node) {
	for _, h := range nodes {
		fmt.Println(virtmap.Info(nodes, h.Name))
	}
}
