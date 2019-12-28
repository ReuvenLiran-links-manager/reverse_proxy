package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const fallbackPort = "9000"

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

func main() {
	logSetup()
	router := mux.NewRouter().UseEncodedPath()

	router.PathPrefix("/{tab}/proxy-service-worker.js").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/proxy-service-worker.js")
	})
	router.PathPrefix("/{tab}/static/proxy-init.html").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/proxy-init.html")
	})
	router.PathPrefix("/static/proxy-register.js").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/proxy-register.js")
	})

	router.HandleFunc("/{tab}/proxy/{url}", handleRequestAndRedirect)
	router.HandleFunc("/proxy/{url}", handleRequestAndRedirect)

	log.Fatal(http.ListenAndServe(getListenAddress(), router))
}
