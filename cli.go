package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/subsonic74/virtmapper/virtmap"
)

func Query(node string) (virtmap.Vmap, error) {
	response, err := http.Get("http://" + httpServer + "/api/v1/" + query)
	if err != nil {
		fmt.Printf("Get() error, %v\n", err)
		return nil, err
	}

	var vmap virtmap.Vmap
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("ReadAll() error, %v\n", err)
		return nil, err
	}
	var genericReply map[string]interface{}
	err = json.Unmarshal(body, &genericReply)
	if err != nil {
		fmt.Printf("JSON Unmarshalling error: %v", err)
		return nil, err
	}

	if data, isQueryError := genericReply["error"]; isQueryError {
		return nil, errors.New(data.(string)+"\n")
	}

	if _, isNode := genericReply["node"]; isNode {
		type Vhost struct {
			Node   virtmap.Node   `json:"node"`
			Guests []virtmap.Node `json:"guests"`
		}
		tmp := Vhost{}
		err = json.Unmarshal(body, &tmp)
		vmap = append(vmap, tmp.Node)
		for _, g := range tmp.Guests {
			vmap = append(vmap, g)
		}
	}

	if _, isVmap := genericReply["vmap"]; isVmap {
		type full struct {
			Vmap   []virtmap.Node `json:"vmap"`
		}
		tmp := full{}
		err = json.Unmarshal(body, &tmp)
		vmap = virtmap.Vmap(tmp.Vmap)
	}

	if err != nil {
		fmt.Printf("Unmarshal() error, %v\n", err)
		return vmap, err
	}
	return vmap, nil
}

func Display(vmap virtmap.Vmap) {
	for _, h := range vmap {
		fmt.Println(vmap.Info(h.Name))
	}
}
