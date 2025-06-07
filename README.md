Load Balancer in Go

This project implements a load balancer in Go, simulating traffic distribution across mock servers with built-in health checks and configurable load-balancing algorithms.

Step 1: Initialize your project in project folder go-load-balancer

> > go mod init go-load-balancer

The go mod init command creates a go.mod file to track your code's dependencies. So far, the file includes only the name of your module and the Go version your code supports. But as you add dependencies, the go.mod file will list the versions your code depends on. This keeps builds reproducible and gives you direct control over which module versions to use.

Step 2: Create the Project Structure

go-load-balancer/
├── mock-servers/
│ └── mock_server.go # Code for creating mock servers
├── load-balancer/
│ └── load_balancer.go # Code for load balancer logic
├── test/
└── test_traffic.go # Code for simulating traffic

Problems Encountered

Problem 1: When running multiple mock servers in the same process, a route conflict error occurred:

PS D:\Projects\go-load-balancer\mock-servers> go run mock_server.go
Starting mock server on port 8083
panic: pattern "/" (registered at D:/Projects/go-load-balancer/mock-servers/mock_server.go:13) conflicts with pattern "/" (registered at D:/Projects/go-load-balancer/mock-servers/mock_server.go:13): mock_server.go:13):
/ matches the same requests as /

goroutine 7 [running]:
net/http.(\*ServeMux).register(...)
C:/Program Files/Go/src/net/http/server.go:2797
net/http.HandleFunc({0x125a008?, 0x144ab70?}, 0x144ab68?)
C:/Program Files/Go/src/net/http/server.go:2791 +0x86
main.startServer({0x125a265, 0x4}, 0xc00004dfb8?)
D:/Projects/go-load-balancer/mock-servers/mock_server.go:13 +0xb9
created by main.main in goroutine 1
D:/Projects/go-load-balancer/mock-servers/mock_server.go:32 +0x8a
exit status 2

This happened because all servers shared the global http.DefaultServeMux for route handling, which doesn't allow multiple registrations of the same route.

Solution: To resolve this, we created a custom ServeMux for each mock server:

- Used http.NewServeMux() to create a new multiplexer.
- Registered routes ("/") to this custom multiplexer instead of the global one.
- Explicitly passed the custom multiplexer to http.ListenAndServe.

Reasoning:
In Go's net/http package, the http.HandleFunc function registers routes (like "/") to a default multiplexer (http.DefaultServeMux). When you call http.ListenAndServe, it uses this global multiplexer by default if you don’t specify your own.

The problem arises because all your mock servers are running in the same process and registering the same "/" routes on the shared default multiplexer. When multiple servers attempt to register the same routes, Go’s HTTP server detects a conflict, as a route can only be registered once per multiplexer.

Why Does This Happen?
Shared Global State: The default multiplexer (http.DefaultServeMux) is shared across the entire application. If multiple http.HandleFunc calls attempt to register the same route, the application throws a conflict error.

Independent Servers: Each mock server should ideally have its own routing logic. The shared multiplexer doesn't allow this separation.

Why This Workaround Fixes the Issue?

1. Creating a New ServeMux for Each Server
   The http.NewServeMux() function creates a new instance of a multiplexer. Each instance manages its own set of route registrations. By doing this:

Each server operates with an independent routing table.

Routes for one server won’t conflict with routes for another server.

2. Explicitly Binding the ServeMux to the Server
   When calling http.ListenAndServe, you can pass the specific ServeMux instance that the server should use. This ensures that the server doesn’t rely on the global DefaultServeMux, avoiding route conflicts.

Why Does the Default ServeMux Exist?
The DefaultServeMux is a convenient tool for simple applications where only one HTTP server is used. It avoids the need to explicitly manage route registration in small projects. However, in applications with multiple servers or complex routing requirements, relying on a shared multiplexer can lead to issues, as seen here.

Problem 2:

Solution:

Reasoning:
