package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/mux"
)

/*
	Structs
*/

var myBase string
var prevBase string
var path string

const fallbackPort = "9000"

type myTransport struct {
	// Uncomment this if you want to capture the transport
	// CapturedTransport http.RoundTripper
}

func (t *myTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	response.Header.Del("X-Frame-Options")         //TODO: Can remove when service and ui will have the same domain
	response.Header.Del("Content-Security-Policy") // TODO: change Content-Security-Policy: upgrade-insecure-requests; frame-ancestors 'self' https://stackexchange.com  to localhost

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
	port := getEnv("PORT", fallbackPort)
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
	req.URL, _ = url.Parse(myURL.String())

	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = myURL.Host

	proxy.Transport = &myTransport{}
	proxy.Director = report
	proxy.ServeHTTP(res, req)
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	myURL := mux.Vars(req)["url"]
	myURL, err := url.PathUnescape(myURL)
	if err != nil {
		log.Println("Error", err)
		return
	}

	URL, err := url.Parse(myURL)
	if err != nil {
		log.Println("Error", err)
		return
	}
	request, err := copyRequest(req, URL)
	if err != nil {
		log.Println("Error", err)
		return
	}
	serveReverseProxy(myURL, res, request)
}

func main() {
	logSetup()
	router := mux.NewRouter().UseEncodedPath()

	router.PathPrefix("/proxy-service-worker.js").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/service-worker1.js")
	})
	router.PathPrefix("/static/proxy-init.html").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/html.html")
	})
	router.PathPrefix("/static/proxy-register.js").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/register.js")
	})
	router.HandleFunc("/proxy/{url}", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(getListenAddress(), router))
}

func copyRequest(request *http.Request, targetURL *url.URL) (*http.Request, error) {
	target := request.URL
	target.Scheme = targetURL.Scheme
	target.Host = targetURL.Host
	target.Path = targetURL.Path
	target.RawPath = targetURL.RawPath
	target, _ = url.Parse(targetURL.String())

	req, err := http.NewRequest(request.Method, target.String(), request.Body)
	if err != nil {
		return nil, err
	}
	for key := range request.Header {
		req.Header.Set(key, request.Header.Get(key))
	}

	if targetURL.Scheme == "https" {
		req.Proto = "HTTP/2"
		req.ProtoMajor = 2
		req.ProtoMinor = 0
	} else if targetURL.Scheme == "http" {
		req.Proto = "HTTP/1.1"
		req.ProtoMajor = 1
		req.ProtoMinor = 1
	}

	req.Header.Set("Referer", target.String())
	req.Header.Set("Host", target.Host)

	// Go supports gzip compression, but not Brotli.
	// Since the underlying transport handles compression, remove this header to avoid problems.
	req.Header.Del("Accept-Encoding")
	return req, nil
}

func report(r *http.Request) {
}
