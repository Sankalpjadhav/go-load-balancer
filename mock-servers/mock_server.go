package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	serverMap = make(map[string]*http.Server) // Tracks running servers
	mutex     sync.Mutex                      // Protects access to serverMap
)

func startServer(port string, wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux() // Create a new ServeMux instance for each server
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received at server on port %s", port)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello from mock server on port %s!", port)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mutex.Lock()
	serverMap[port] = server
	mutex.Unlock()

	fmt.Printf("Starting mock server on port %s\n", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server on port %s failed: %s\n", port, err)
	}

	mutex.Lock()
	delete(serverMap, port)
	mutex.Unlock()
	fmt.Printf("Server on port %s stopped\n", port)
}

func stopServer(port string) {
	mutex.Lock()
	server, exists := serverMap[port]
	mutex.Unlock()

	if !exists {
		fmt.Printf("No server running on port %s.\n", port)
		return
	}

	if err := server.Close(); err != nil {
		log.Printf("Error stopping server on port %s: %s\n", port, err)
	} else {
		fmt.Printf("Server on port %s has been stopped.\n", port)
	}
}

func main() {
	var wg sync.WaitGroup

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Mock Server Management CLI")
	fmt.Println("Type 'start <port>' to start a new server or 'stop <port>' to stop an existing server.")
	fmt.Println("Example: start 8081 or stop 8082")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		command := scanner.Text()
		parts := strings.Fields(command)

		if len(parts) != 2 {
			fmt.Println("Invalid command. Please use 'start <port>' or 'stop <port>'.")
			continue
		}

		action, port := parts[0], parts[1]

		switch action {
		case "start":
			mutex.Lock()
			_, exists := serverMap[port]
			mutex.Unlock()

			if exists {
				fmt.Printf("Server on port %s is already running.\n", port)
			} else {
				wg.Add(1)
				go startServer(port, &wg)
			}
		case "stop":
			stopServer(port)
		default:
			fmt.Println("Invalid action. Use 'start' or 'stop'.")
		}
	}

	wg.Wait()
}
