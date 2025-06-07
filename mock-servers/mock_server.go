package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func startServer(port string, wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux() // Create a new ServeMux instance for each server

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from mock server on port %s!", port)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	fmt.Printf("Starting mock server on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux)) // Use the server-specific mux
}

func main() {
	var wg sync.WaitGroup
	ports := []string{"8081", "8082", "8083"}

	for _, port := range ports {
		wg.Add(1)
		go startServer(port, &wg)
	}

	wg.Wait()
}
