package main

import (
	"fmt"
	"log"
	"net/http"
)

// handler function to write response
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check to only respond to requests for the root path "/"
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Set the content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write a simple HTML response
	fmt.Fprintf(w, "<h1>Hello World from Go!</h1>")
}

func main() {
	// Register the handler function for the "/" route
	http.HandleFunc("/", helloHandler)

	// Define the port the server will listen on
	port := "8080"
	fmt.Printf("Starting server on port %s...\n", port)
	fmt.Printf("Visit http://localhost:%s in your browser.\n", port)

	// Start the HTTP server
	// ListenAndServe blocks until the server is stopped (or crashes)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err) // Log error if server fails
	}
}
