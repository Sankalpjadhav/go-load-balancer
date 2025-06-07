/*
1.) Define the load balancer's structure.
2.) Accept client requests.
3.) Distribute requests using the Round Robin algorithm.
*/

package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type LoadBalancer struct {
	Servers       []string       // List of mock server addresses
	Index         uint32         // Tracks the next server to use
	HealthChecker *HealthChecker // Reference to health checker
}

// Round Robin algorithm
func (lb *LoadBalancer) getNextServer() string {
	lb.HealthChecker.Mutex.Lock()
	defer lb.HealthChecker.Mutex.Unlock()

	var activeServers []string
	for _, server := range lb.Servers {
		if lb.HealthChecker.Active[server] {
			activeServers = append(activeServers, server)
		}
	}

	if len(activeServers) == 0 {
		log.Println("No healthy servers available!")
		return ""
	}

	index := atomic.AddUint32(&lb.Index, 1) // Increment the index by 1 for the next mock server
	return activeServers[int(index)%len(activeServers)]
}

// Handle client requests
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := lb.getNextServer()
	targetURL := fmt.Sprintf("http://%s%s", server, r.URL.Path)

	// Forward the request to the selected server
	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Server not reachable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Forward the response back to the client
	w.WriteHeader(resp.StatusCode)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	_, _ = w.Write([]byte(fmt.Sprintf("Response from server: %s", server)))
}

func main() {
	// list of mock servers
	servers := []string{
		"localhost:8081",
		"localhost:8082",
		"localhost:8083",
	}

	// Initialize the health checker
	hc := &HealthChecker{
		Servers:  servers,
		Active:   make(map[string]bool),
		Interval: 15 * time.Second,
	}

	// Start the health checker in a separate goroutine
	go hc.Start()

	// Create the load balancer
	lb := &LoadBalancer{
		Servers:       servers,
		HealthChecker: hc,
	}

	// Start the load balancer server
	fmt.Println("Starting load balancer on port 9090...")
	log.Fatal(http.ListenAndServe(":9090", lb))
}
