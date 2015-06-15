package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/subsonic74/virtmapper/virtmap"
)

func Query(node string) virtmap.Vmap {
	response, err := http.Get("http://" + httpServer + "/api/v1/" + query)
	if err != nil {
		fmt.Printf("Get() error, %s\n", err.Error())
		return nil
	}
	var vmap virtmap.Vmap
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

func Display(vmap virtmap.Vmap) {
	for _, h := range vmap {
		fmt.Println(vmap.Info(h.Name))
	}
}
