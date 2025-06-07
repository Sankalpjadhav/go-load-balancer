/*
1.) Define the load balancer's structure.
2.) Accept client requests.
3.) Distribute requests using the Round Robin algorithm.
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LoadBalancer struct {
	Servers       map[string]bool // Tracks server addresses and their statuses
	Index         uint32          // Tracks the next server to use
	HealthChecker *HealthChecker  // Reference to health checker
	Mutex         sync.Mutex      // Protects access to Servers
}

// Add a server to the load balancer
func (lb *LoadBalancer) AddServer(server string) {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	if _, exists := lb.Servers[server]; exists {
		fmt.Printf("Server %s is already in the pool.\n", server)
	} else {
		lb.Servers[server] = true
		lb.HealthChecker.AddServer(server)
		fmt.Printf("Server %s added to the pool.\n", server)
	}
}

// Remove a server from the load balancer
func (lb *LoadBalancer) RemoveServer(server string) {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	if _, exists := lb.Servers[server]; exists {
		delete(lb.Servers, server)
		lb.HealthChecker.RemoveServer(server)
		fmt.Printf("Server %s removed from the pool.\n", server)
	} else {
		fmt.Printf("Server %s is not in the pool.\n", server)
	}
}

// Round Robin algorithm
func (lb *LoadBalancer) getNextServer() string {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	var activeServers []string
	for server := range lb.Servers {
		// Check health status using HealthChecker
		if lb.HealthChecker.IsServerHealthy(server) {
			activeServers = append(activeServers, server)
		}
	}

	if len(activeServers) == 0 {
		log.Println("No healthy servers available!")
		return ""
	}

	index := atomic.AddUint32(&lb.Index, 1) // Increment the index by 1 for the next server
	return activeServers[int(index)%len(activeServers)]
}

// Handle client requests
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := lb.getNextServer()
	if server == "" {
		http.Error(w, "No servers available", http.StatusServiceUnavailable)
		return
	}

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
	// Initialize the load balancer
	lb := &LoadBalancer{
		Servers:       make(map[string]bool),
		HealthChecker: &HealthChecker{Interval: 15 * time.Second, Active: make(map[string]bool)},
	}

	// Start the health checker
	go lb.HealthChecker.Start()

	// Start the load balancer server
	go func() {
		fmt.Println("Starting load balancer on port 9090...")
		log.Fatal(http.ListenAndServe(":9090", lb))
	}()

	// CLI for managing servers
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Load Balancer Management CLI")
	fmt.Println("Type 'add <server>' to add a server or 'remove <server>' to remove a server.")
	fmt.Println("Example: add localhost:8081 or remove localhost:8082")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		command := scanner.Text()
		parts := strings.Fields(command)

		if len(parts) != 2 {
			fmt.Println("Invalid command. Use 'add <server>' or 'remove <server>'.")
			continue
		}

		action, server := parts[0], parts[1]

		switch action {
		case "add":
			lb.AddServer(server)
		case "remove":
			lb.RemoveServer(server)
		default:
			fmt.Println("Invalid action. Use 'add' or 'remove'.")
		}
	}
}
