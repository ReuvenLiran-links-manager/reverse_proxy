package main

import (
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

const fallback_port = "9000"

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
	if err != nil {
		panic(err)
	}
	if response == nil {
		return response, err
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
	port := getEnv("PORT", fallback_port)
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

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	redisClient := GetRedisClient()

	myURL := redisClient.GetHost()

	if myURL != redisClient.GetPrevHost() {
		req.URL.Path = path
		redisClient.SetPrevHost(myURL)
	} else {
		req.URL.Path, _ = url.PathUnescape(req.URL.Path[1:len(req.URL.Path)])
	}

	serveReverseProxy(myURL, res, req)
}

func rewriter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Simple URL rewriter. Rewrite if it's started with API path
		pathReq := r.RequestURI
		if strings.HasPrefix(pathReq, "/new") {

			pe := strings.TrimLeft(pathReq, "/new")
			a, _ := url.PathUnescape(pe)
			myURL, _ := url.Parse(a)
			myBase = myURL.Scheme + "://" + myURL.Host

			redisClient := GetRedisClient()
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

func main() {
	logSetup()
	r := mux.NewRouter()
	r.HandleFunc("/{url}", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(getListenAddress(), rewriter(r)))
}
