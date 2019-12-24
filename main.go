package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	// response, err := http.DefaultTransport.RoundTrip(request)
	response, err := http.DefaultTransport.RoundTrip(request)

	// targetURL := request.URL.String()

	response.Header.Del("X-Frame-Options")
	request.Header.Del("X-Frame-Options")
	// err = injectJS(response, targetURL)
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
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = myURL.Host
	// if len(req.URL.Path) > 1 {
	// 	req.URL.Path = req.URL.Path[1:len(req.URL.Path)]
	// }

	proxy.Transport = &myTransport{}
	proxy.Director = report
	proxy.ServeHTTP(res, req)
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	// redisClient := GetRedisClient()
	myURL := mux.Vars(req)["url"]
	myURL, err := url.PathUnescape(myURL)
	if err != nil {
		log.Println("Error", err)
		return
	}

	// myURL := redisClient.GetHost()

	// if myURL != redisClient.GetPrevHost() {
	// 	req.URL.Path = path
	// 	redisClient.SetPrevHost(myURL)
	// } else {
	// 	req.URL.Path, _ = url.PathUnescape(req.URL.Path[1:len(req.URL.Path)])
	// }

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

// func rewriter(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		//Simple URL rewriter. Rewrite if it's started with API path
// 		pathReq := r.RequestURI
// 		if strings.HasPrefix(pathReq, "/new") {

// 			pe := strings.TrimLeft(pathReq, "/new")
// 			a, _ := url.PathUnescape(pe)
// 			myURL, _ := url.Parse(a)
// 			myBase = myURL.Scheme + "://" + myURL.Host

// 			redisClient := GetRedisClient()
// 			redisClient.SetHost(myBase)
// 			redisClient.SetPrevHost("")

// 			path = myURL.Path
// 			prevBase = ""
// 			r.URL.Path = "/" + pe
// 			r.URL.RawQuery = ""
// 		} else {
// 			r.URL.Path = "/" + url.PathEscape(r.URL.Path)
// 		}
// 		if myBase != "" {
// 			h.ServeHTTP(w, r)
// 		}
// 	})
// }

func main() {
	logSetup()
	router := mux.NewRouter().UseEncodedPath()

	// fs := http.FileServer(http.Dir("static"))
	// router.Handle("/public", fs)
	// router.PathPrefix("/").Handler(fs)
	// router.PathPrefix("/public/").Handler(http.FileServer(http.Dir("./static/")))
	// s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	// router.Pa0Prefix("/static/").Handler(s)

	router.PathPrefix("/service-worker.js").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/service-worker1.js")
	})
	router.PathPrefix("/html.html").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/html.html")
	})
	router.PathPrefix("/register.js").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/register.js")
	})
	router.HandleFunc("/{url}", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(getListenAddress(), router))
}

func copyRequest(request *http.Request, targetURL *url.URL) (*http.Request, error) {
	target := request.URL
	target.Scheme = targetURL.Scheme
	target.Host = targetURL.Host
	target.Path = targetURL.Path
	// if len(targetURL.Path) > 1 {
	// 	target.Path = targetURL.Path[1:len(targetURL.Path)]
	// }
	target.RawPath = targetURL.RawPath

	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: *sourceInsecure}

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

	// Don't let a stray referer header give away the location of our site.
	// Note that this will not prevent leakage from full URLs.
	// if request.Referer() != "" {
	// req.Header.Set("Referer", strings.Replace(request.Referer(), request.Host, targetURL.Host, -1))
	// }

	req.Header.Set("Referer", target.String())
	req.Header.Set("Host", target.Host)

	// Go supports gzip compression, but not Brotli.
	// Since the underlying transport handles compression, remove this header to avoid problems.
	req.Header.Del("Accept-Encoding")
	req.Header.Del("X-Frame-Options")
	return req, nil
}

// Transform Injects JavaScript into an HTML response.
func injectJS(response *http.Response, targetURL string) error {
	if !strings.Contains(response.Header.Get("Content-Type"), "text/html") {
		return nil
	}

	// Prevent NewDocumentFromReader from closing the response body.
	responseText, err := ioutil.ReadAll(response.Body)
	responseBuffer := bytes.NewBuffer(responseText)
	response.Body = ioutil.NopCloser(responseBuffer)
	if err != nil {
		return err
	}

	document, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return err
	}

	// href := "https://www.google.com?t=123" //+ url.PathEscape(targetURL)
	// href := "http://localhost:3000"
	// payload := fmt.Sprintf("<base href='%s'></base>", href)

	// serviceWorker := "http://localhost:9000/register.js"
	// serviceWorkerTag := fmt.Sprintf("<script src='%s'></script>", serviceWorker)

	selection := document.
		Find("head").
		// AppendHtml(payload).
		// AppendHtml(serviceWorkerTag).
		Parent()

	html, err := selection.Html()
	if err != nil {
		return err
	}
	response.Body = ioutil.NopCloser(bytes.NewBufferString(html))
	return nil
}

func report(r *http.Request) {
	// r.Host = "stackoverflow.com"
	// r.URL.Host = "stackoverflow.com"
	// r.URL.Scheme = "https"

	// r.Host = r.Host
	// r.URL.Host = r.Host
	// r.URL.Scheme = r.URL.Scheme
}
