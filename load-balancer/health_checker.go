/*
1.) Periodically check the /health endpoint of each mock server.
2.) Update the load balancer’s list of active servers.
3.) Allow dynamic addition and removal of servers.
*/
package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type HealthChecker struct {
	Active   map[string]bool // Tracks the health status of each server
	Mutex    sync.Mutex      // Protects access to the Active map
	Interval time.Duration   // Interval for periodic health checks
	stopCh   chan struct{}   // Channel to stop the health checker
	wg       sync.WaitGroup  // WaitGroup for goroutine management
}

// Check the health of each server
func (hc *HealthChecker) checkHealth() {
	hc.Mutex.Lock()
	defer hc.Mutex.Unlock()

	for server := range hc.Active {
		go func(s string) {
			resp, err := http.Get("http://" + s + "/health")
			hc.Mutex.Lock()
			defer hc.Mutex.Unlock()
			if err != nil || resp.StatusCode != http.StatusOK {
				hc.Active[s] = false
				log.Printf("Server %s is down", s)
			} else {
				hc.Active[s] = true
				log.Printf("Server %s is healthy", s)
			}
		}(server)
	}
}

func (hc *HealthChecker) IsServerHealthy(server string) bool {
	hc.Mutex.Lock()
	defer hc.Mutex.Unlock()

	healthy, exists := hc.Active[server]
	return exists && healthy
}

// Start periodic health checks
func (hc *HealthChecker) Start() {
	hc.stopCh = make(chan struct{}) // Create a new stop channel
	hc.wg.Add(1)

	go func() {
		defer hc.wg.Done()
		ticker := time.NewTicker(hc.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.checkHealth()
			case <-hc.stopCh:
				log.Println("Health checker stopped.")
				return
			}
		}
	}()
}

// Stop the health checker
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
	hc.wg.Wait() // Wait for all goroutines to finish
}

// Add a server to the health checker
func (hc *HealthChecker) AddServer(server string) {
	hc.Mutex.Lock()
	defer hc.Mutex.Unlock()

	if _, exists := hc.Active[server]; !exists {
		hc.Active[server] = true // Assume server is healthy initially
		log.Printf("Server %s added to health checker", server)
	}
}

// Remove a server from the health checker
func (hc *HealthChecker) RemoveServer(server string) {
	hc.Mutex.Lock()
	defer hc.Mutex.Unlock()

	if _, exists := hc.Active[server]; exists {
		delete(hc.Active, server)
		log.Printf("Server %s removed from health checker", server)
	}
}
