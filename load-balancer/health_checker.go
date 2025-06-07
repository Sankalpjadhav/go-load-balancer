/*
1.) Periodically check the /health endpoint of each mock server.
2.) Update the load balancer’s list of active servers.
*/
package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type HealthChecker struct {
	Servers  []string
	Active   map[string]bool
	Mutex    sync.Mutex
	Interval time.Duration
}

// Check the health of each server
func (hc *HealthChecker) checkHealth() {
	for _, server := range hc.Servers {
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

// Start periodic health checks
func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.Interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		hc.checkHealth()
	}
}
