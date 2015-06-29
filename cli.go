package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/subsonic74/virtmapper/virtmap"
)

var HTTPGetter = func(url string) (*http.Response, error) {
	return http.Get(url)
}

func Query(query string) (virtmap.Vmap, error) {
	response, err := HTTPGetter("http://" + httpServer + "/api/v1/" + query)
	if err != nil {
		fmt.Printf("Get() error, %v\n", err)
		return virtmap.Vmap{}, err
	}

	var vmap virtmap.Vmap
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("ReadAll() error, %v\n", err)
		return virtmap.Vmap{}, err
	}

	var genericReply map[string]interface{}
	err = json.Unmarshal(body, &genericReply)
	if err != nil {
		fmt.Printf("JSON Unmarshalling error: %v", err)
		return virtmap.Vmap{}, err
	}

	if data, isQueryError := genericReply["error"]; isQueryError {
		return virtmap.Vmap{}, errors.New(data.(string) + "\n")
	} else {
		err = json.Unmarshal(body, &vmap)
	}

	if err != nil {
		fmt.Printf("Unmarshal() error, %v\n", err)
		return vmap, err
	}
	return vmap, nil
}

func Display(vmap virtmap.Vmap) {
	for n, _ := range vmap.Hosts {
		fmt.Println(vmap.Info(n))
	}
	for n, _ := range vmap.Guests {
		fmt.Println(vmap.Info(n))
	}
}
