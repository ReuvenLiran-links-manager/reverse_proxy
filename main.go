package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

/*
	Structs
*/

var myBase string
var prevBase string
var path string
var redisClient Redis

const FALLBACK_PORT = "9000"

type myTransport struct {
	// Uncomment this if you want to capture the transport
	// CapturedTransport http.RoundTripper
}

func (t *myTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	// or, if you captured the transport
	// response, err := t.CapturedTransport.RoundTrip(request)

	// The httputil package provides a DumpResponse() func that will copy the
	// contents of the body into a []byte and return it. It also wraps it in an
	// ioutil.NopCloser and sets up the response to be passed on to the client.
	if response == nil {
		return nil, nil
	}

	if response.StatusCode == 404 {
		client := &http.Client{}
		fallbackURL := os.Getenv("FALLBACK_URL")
		myURL, _ := url.Parse(fallbackURL)

		request.URL.Scheme = myURL.Scheme
		request.URL.Host = myURL.Host
		request.URL.Path = myURL.Path + request.URL.Path
		request.RequestURI = ""
		clResp, clErr := client.Do(request)
		return clResp, clErr
	}

	return response, err
}

/*
	Utilities
*/

// Get env var or default
func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	return fallback
}

/*
	Getters
*/

// Get the port to listen on
func getListenAddress() string {
	port := getEnv("PORT", FALLBACK_PORT)
	return ":" + port
}

/*
	Logging
*/

// Log the env variables required for a reverse proxy
func logSetup() {
	log.Printf("Server will run on: %s\n", getListenAddress())
}

/*
	Reverse Proxy Logic
*/

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	newTarget, _ := url.PathUnescape(target)

	myURL, _ := url.Parse(newTarget)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(myURL)

	// Update the headers to allow for SSL redirection
	req.URL.Host = myURL.Host
	req.URL.Scheme = myURL.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = myURL.Host
	proxy.Transport = &myTransport{}
	proxy.ServeHTTP(res, req)
}

// Get a json decoder for a given requests body
func requestBodyDecoder(request *http.Request) *json.Decoder {
	// Read body to buffer
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		panic(err)
	}

	// Because go lang is a pain in the ass if you read the body then any susequent calls
	// are unable to read the body again....
	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body)))
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	myURL := redisClient.GetHost()

	// Update the headers to allow for SSL redirection
	if myURL != redisClient.GetPrevHost() {
		req.URL.Path = path
		redisClient.SetPrevHost(myURL)
	} else {
		req.URL.Path, _ = url.PathUnescape(req.URL.Path[1:len(req.URL.Path)])
	}

	serveReverseProxy(myURL, res, req)
}

func Rewriter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Simple URL rewriter. Rewrite if it's started with API path
		pathReq := r.RequestURI
		if strings.HasPrefix(pathReq, "/new") {
			//Use url.QueryEscape for pre go1.8
			pe := strings.TrimLeft(pathReq, "/new")
			a, _ := url.PathUnescape(pe)
			myURL, _ := url.Parse(a)
			myBase = myURL.Scheme + "://" + myURL.Host

			redisClient.SetHost(myBase)
			redisClient.SetPrevHost("")

			path = myURL.Path
			prevBase = ""
			r.URL.Path = "/" + pe
			r.URL.RawQuery = ""
		} else {
			r.URL.Path = "/" + url.PathEscape(r.URL.Path)
		}
		if myBase != "" {
			h.ServeHTTP(w, r)
		}
	})
}

// func ExampleNewClient() {
// 	client := redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // no password set
// 		DB:       0,  // use default DB
// 	})

// 	pong, err := client.Ping().Result()
// 	log.Printf(pong, err)
// 	// Output: PONG <nil>
// 	err = client.Set("key", "value", 0).Err()
// 	if err != nil {
// 		panic(err)
// 	}

// 	val, err := client.Get("key").Result()
// 	if err != nil {
// 		panic(err)
// 	}
// 	log.Printf("key", val)

// 	val2, err := client.Get("key2").Result()
// 	if err == redis.Nil {
// 		log.Printf("key2 does not exist")
// 	} else if err != nil {
// 		panic(err)
// 	} else {
// 		log.Printf("key2", val2)
// 	}
// 	// Output: key value
// 	// key2 does not exist
// }

func main() {
	logSetup()
	redisClient.ConnectToRedis()
	// redisClient.SetHost("ssssssss")
	// log.Printf(redisClient.GetHost())

	r := mux.NewRouter()
	r.HandleFunc("/{url}", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(getListenAddress(), Rewriter(r)))
}
