# Custom Load Balancer in Go

This project implements a load balancer in Go, simulating traffic distribution across mock servers with built-in health checks and configurable load-balancing algorithms.

## Project Structure

```
go-load-balancer/
├── mock-servers/
│ └── mock_server.go # Code for creating mock servers
├── load-balancer/
│ └── load_balancer.go # Code for load balancer logic
| └── health_checker.go # Code for checking health of mock servers
├── test/
├── go.mod
└── README.md
```

## Prerequisites

- Install Go (version 1.23 or later).
- Set up the project by initializing the Go module.

```
go mod init go-load-balancer
```

The go mod init command creates a go.mod file to track your code's dependencies. So far, the file includes only the name of your module and the Go version your code supports. But as you add dependencies, the go.mod file will list the versions your code depends on. This keeps builds reproducible and gives you direct control over which module versions to use.

## Steps to Run

> [!NOTE]
> Open three terminals to see the logs for Client Requests, Load Balancer, and Mock Servers.

### Step 1: Start Mock Servers [Terminal 1]

The mock servers simulate backend servers that the load balancer will forward requests to. Each server listens on a unique port and responds to health checks and client requests.

1. Change to the mock-servers directory:

```
cd mock-servers
```

2. Run the mock servers:

```
go run mock_server.go
```

3. You should see output like:

```
Mock Server Management CLI
Type 'start <port>' to start a new server or 'stop <port>' to stop an existing server.
Example: start 8081 or stop 8082
>
```

4. Start/Stop mock servers as per your needs:

```
start 8081
start 8082
```

### Step 2: Start the Load Balancer [Terminal 2]

The load balancer distributes client requests to the mock servers using a round-robin algorithm. It also periodically checks the health of servers to ensure requests are not routed to unhealthy servers.

1. Change to the load-balancer directory:

```
cd ../load-balancer
```

2. Run the Load Balancer and Health Checker

```
go run .
```

3. You should see output like:

```
Load Balancer Management CLI
Type 'add <server>' to add a server or 'remove <server>' to remove a server.
Example: add localhost:8081 or remove localhost:8082
> Starting load balancer on port 9090...
```

4. Add the mock servers you already started as part of `Step 1`, these servers will be tracked by the Load Balancer

```
add localhost:8081
add localhost:8082
```

### Step 3: Send Client Requests [Terminal 3]

Once the load balancer is running, you can simulate client traffic by sending HTTP requests to the load balancer's port (9090).

1. Use curl to send requests:

```
curl -s http://localhost:9090
```

2. You will see responses like:

```
Response from server: localhost:8081
```

`Note:` The load balancer will forward requests to different servers in a round-robin manner for the subsequent requests.

### Step 4: Simulate Server Failures

You can stop a mock server to simulate a failure and observe how the load balancer responds.

1. Stop a mock server by running below command on `Terminal 1`

```
stop 8081
```

2. The load balancer will detect the server is down via health checks and stop routing requests to it.

3. Logs in the load balancer will show:

```
Server localhost:8081 is down
Server localhost:8082 is healthy
```

4. Verify by sending requests again. The load balancer will route traffic only to healthy servers.

## Problems Encountered

### Problem 1: When running multiple mock servers in the same process, a route conflict error occurred:

```
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
```

This happened because all servers shared the global http.DefaultServeMux for route handling, which doesn't allow multiple registrations of the same route.

**Solution:**

To resolve this, I created a custom ServeMux for each mock server:

- Used http.NewServeMux() to create a new multiplexer.
- Registered routes ("/") to this custom multiplexer instead of the global one.
- Explicitly passed the custom multiplexer to http.ListenAndServe.

**Reasoning:**

In Go's **net/http** package, the http.HandleFunc function registers routes (like "/") to a default multiplexer (http.DefaultServeMux). When you call http.ListenAndServe, it uses this global multiplexer by default if you don’t specify your own.

The problem arises because all your mock servers are running in the same process and registering the same "/" routes on the shared default multiplexer. When multiple servers attempt to register the same routes, Go’s HTTP server detects a conflict, as a route can only be registered once per multiplexer.

**Why Does This Happen?**

Shared Global State: The default multiplexer (http.DefaultServeMux) is shared across the entire application. If multiple http.HandleFunc calls attempt to register the same route, the application throws a conflict error.

Independent Servers: Each mock server should ideally have its own routing logic. The shared multiplexer doesn't allow this separation.

**Why This Workaround Fixes the Issue?**

1. Creating a New ServeMux for Each Server

   The http.NewServeMux() function creates a new instance of a multiplexer. Each instance manages its own set of route registrations. By doing this:
   Each server operates with an independent routing table.
   Routes for one server won’t conflict with routes for another server.

2. Explicitly Binding the ServeMux to the Server

   When calling http.ListenAndServe, you can pass the specific ServeMux instance that the server should use. This ensures that the server doesn’t rely on the global DefaultServeMux, avoiding route conflicts.

**Why Does the Default ServeMux Exist?**

The DefaultServeMux is a convenient tool for simple applications where only one HTTP server is used. It avoids the need to explicitly manage route registration in small projects. However, in applications with multiple servers or complex routing requirements, relying on a shared multiplexer can lead to issues, as seen here.
