package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// HTTPGetter is a helper func for http.Get(), factored out so it can be a hook for testing.
var HTTPGetter = func(url string) (*http.Response, error) {
	return http.Get(url)
}

// Query is the CLI client function.  Queries the given server for the given host,
// unmarshalls the JSON, and returns a result Vmap.
func Query(httpServer string, query string) (Vmap, error) {
	rawResponse, err := HTTPGetter("http://" + httpServer + apiPrefix + query)
	if err != nil {
		fmt.Printf("Get() error, %v\n", err)
		return Vmap{}, err
	}

	var vmap Vmap
	body, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		fmt.Printf("ReadAll() error, %v\n", err)
		return Vmap{}, err
	}

	var decodedResponse map[string]interface{}
	err = json.Unmarshal(body, &decodedResponse)
	if err != nil {
		fmt.Printf("JSON Unmarshalling error: %v\n", err)
		return Vmap{}, err
	}

	// If the response contains a top-level error, return it
	if data, ok := decodedResponse["error"]; ok {
		return Vmap{}, errors.New(data.(string))
	}
	err = json.Unmarshal(body, &vmap)

	if err != nil {
		fmt.Printf("Unmarshal() error, %v\n", err)
		return vmap, err
	}
	return vmap, nil
}

// Display takes a result Vmap from Query() and displays it to the user.
func Display(vmap Vmap) {
	for n := range vmap.Hosts {
		fmt.Println(vmap.Info(n))
	}
	for n := range vmap.Guests {
		fmt.Println(vmap.Info(n))
	}
}
