package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func getResp(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(body)
	return bodyString
}

// TestProxy - test proxy simple functionality
func TestProxy(t *testing.T) {
	t.Log("TestProxy - test proxy simple functionality")

	exampleString := getResp("http://example.com/")
	proxyString := getResp("http://localhost:9000/tab1/proxy/http%3A%2F%2Fexample.com%2F")
	if proxyString != exampleString {
		t.Error("requests' HTML are not equal")
	}
	// log.Print(exampleString)
	// log.Print(proxyString)
	// t.Errorf("Sum was incorrect, got: %d, want: %d.", 5, 10)
}
