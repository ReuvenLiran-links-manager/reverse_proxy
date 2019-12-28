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

type myTransport struct {
	// Uncomment this if you want to capture the transport
	// CapturedTransport http.RoundTripper
}

// RoundTrip - get response of request and delete header
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

// Serve a reverse proxy for a given url
func serveReverseProxy(targetURL *url.URL, res http.ResponseWriter, req *http.Request) {
	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Update the headers to allow for SSL redirection
	// Set url o request
	req.URL = targetURL

	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = targetURL.Host

	proxy.Transport = &myTransport{}
	proxy.Director = report
	proxy.ServeHTTP(res, req)
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	urlParam := mux.Vars(req)["url"]
	urlParam, err := url.PathUnescape(urlParam)
	if err != nil {
		log.Println("Error", err)
		return
	}

	urlObject, err := url.Parse(urlParam)
	if err != nil {
		log.Println("Error", err)
		return
	}
	request, err := copyRequest(req, urlObject)

	if err != nil {
		log.Println("Error", err)
		return
	}
	serveReverseProxy(urlObject, res, request)
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
