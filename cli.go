package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/subsonic74/virtmapper/virtmap"
)

func Query(host string) []virtmap.Host {
	response, err := http.Get("http://" + httpServer + "/api/v1/" + query)
	if err != nil {
		fmt.Printf("Get() error, %s\n", err.Error())
		return nil
	}
	var vmap []virtmap.Host
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("ReadAll() error, %s\n", err.Error())
		return nil
	}
	err = json.Unmarshal(body, &vmap)
	if err != nil {
		fmt.Printf("Unmarshal() error, %s\n", err.Error())
		return nil
	}
	return vmap
}

func Display(vmap []virtmap.Host) {
	for _, h := range vmap {
		fmt.Println(virtmap.Info(vmap, h.Name))
	}
}
