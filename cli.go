package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Helper func for http.Get(), factored out so it can be a hook for testing.
var HTTPGetter = func(url string) (*http.Response, error) {
	return http.Get(url)
}

// The main CLI client code.  Queries the given server for the given host,
// unmarshalls the JSON, and returns a result Vmap.
func Query(httpServer string, query string) (Vmap, error) {
	response, err := HTTPGetter("http://" + httpServer + "/api/v1/" + query)
	if err != nil {
		fmt.Printf("Get() error, %v\n", err)
		return Vmap{}, err
	}

	var vmap Vmap
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("ReadAll() error, %v\n", err)
		return Vmap{}, err
	}

	var genericReply map[string]interface{}
	err = json.Unmarshal(body, &genericReply)
	if err != nil {
		fmt.Printf("JSON Unmarshalling error: %v\n", err)
		return Vmap{}, err
	}

	if data, isQueryError := genericReply["error"]; isQueryError {
		return Vmap{}, errors.New(data.(string) + "\n")
	} else {
		err = json.Unmarshal(body, &vmap)
	}

	if err != nil {
		fmt.Printf("Unmarshal() error, %v\n", err)
		return vmap, err
	}
	return vmap, nil
}

// Takes a result Vmap from Query() and displays it to the user.
func Display(vmap Vmap) {
	for n, _ := range vmap.Hosts {
		fmt.Println(vmap.Info(n))
	}
	for n, _ := range vmap.Guests {
		fmt.Println(vmap.Info(n))
	}
}
