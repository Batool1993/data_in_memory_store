# In-Memory Data Store CLI & HTTP SDK

This project provides:

- **HTTP server**: an in-memory key-value and list store with TTL support and token-based authentication
- **Go SDK (Client)**: a `StoreClient` interface and `Client` implementation for easy integration
- **CLI**: a command-line interface to interact with the store
- **Middleware**: logging, recovery, and Bearer token authentication

---

## Table of Contents

1. [Features](#features)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Running the Server](#running-the-server)
6. [Using the CLI](#using-the-cli)
7. [Using the Go SDK](#using-the-go-sdk)
8. [Curl Examples](#curl-examples)
9. [Testing](#testing)
10. [Notes & Extensions](#notes--extensions)

---

## Features

- **String operations**: `SetString`, `GetString`, `DeleteString` with TTL
- **List operations**: `LPush`, `RPop` (examples; more can be added)
- **TTL eviction**: background goroutine removes expired entries
- **Token Auth**: `Authorization: Bearer <token>` enforced by middleware
- **Plain-text errors**: server returns HTTP status ≥400 with plain-text messages
- **Logging & Recovery** middleware

---

## Prerequisites

- Go 1.21+
- Docker & Docker Compose (optional, for containerized deployment)

---

## Installation

```bash
git clone https://github.com/Batool1993/data_in_memory_store.git
cd data_in_memory_store
go mod download
```

---

## Configuration

Create a `.env` file in the project root (or export these variables):

```dotenv
STORE_SERVER=http://localhost:8080
STORE_DEFAULT_TTL=60s
CLEANUP_INTERVAL=300s
STORE_API_TOKEN=my-secret-token
```

---

## Running the Server

```bash
# from project root
go run ./cmd/server
```

The server listens on port 8080 by default.

---

## Using the CLI

Build the CLI executable (the `main.go` for the CLI lives in `cmd/`):

```bash
go build -o ds-cli ./cmd
```

> **Note:**
> - `set`, `del`, and `lpush` return exit code 0 on success with no stdout
> - `get` and `rpop` print the retrieved value

```bash
# Set a key with 30s TTL (no output)
./ds-cli --action=set --key=foo --value=bar --ttl=30s

# Get the key (prints the value)
./ds-cli --action=get --key=foo
# → bar

# Delete the key (no output)
./ds-cli --action=del --key=foo

# Left-push items (no output)
./ds-cli --action=lpush --key=mylist --values=a,b,c

# Right-pop (prints the value)
./ds-cli --action=rpop --key=mylist
# → c
```

---
# Docker
Build the image:

docker build -t data-storage:latest .


Run a container:
docker run --rm \
-e STORE_SERVER=http://localhost:8080 \
-e STORE_DEFAULT_TTL=60s \
-e CLEANUP_INTERVAL=300s \
-e STORE_API_TOKEN=my-secret-token \
-p 8080:8080 \
data-storage:latest

You should see:
listening on :8080

Smoke-test the API:
# Set a key
curl -i -X POST http://localhost:8080/v1/string/foo \
-H "Content-Type: application/json" \
-H "Authorization: Bearer my-secret-token" \
-d '{"value":"bar","ttl_seconds":5}'

# Get the key
curl -i http://localhost:8080/v1/string/foo \
-H "Authorization: Bearer my-secret-token"


# Using Docker Compose
run : docker-compose up --build



## Using the Go SDK

Import and use the `client` package in your Go applications:

```go
import (
"context"
"log"
"time"

"github.com/Batool1993/data_in_memory_store/client"
)

func main() {
cli, err := client.NewClient("http://localhost:8080", "my-secret-token")
if err != nil {
log.Fatalf("client init error: %v", err)
}
ctx := context.Background()

// String ops:
if err := cli.SetString(ctx, "foo", "bar", 10*time.Second); err != nil {
log.Fatalf("SetString error: %v", err)
}
val, err := cli.GetString(ctx, "foo")
if err != nil {
log.Fatalf("GetString error: %v", err)
}
log.Printf("foo = %q", val)
if err := cli.DeleteString(ctx, "foo"); err != nil {
log.Fatalf("DeleteString error: %v", err)
}

// List ops:
// Push items ["a","b","c"] onto the left of "mylist"
if err := cli.LPush(ctx, "mylist", "a", "b", "c"); err != nil {
log.Fatalf("LPush error: %v", err)
}
// Pop one item off the right of "mylist"
item, err := cli.RPop(ctx, "mylist")
if err != nil {
log.Fatalf("RPop error: %v", err)
}
log.Printf("RPop mylist → %q", item)
}

```

> **Tip:** All SDK methods accept a `context.Context` for timeouts, cancellations, or carrying request-scoped metadata.

---

## Curl Examples

```bash
# Set with 5s TTL
curl -i -X POST http://localhost:8080/v1/string/foo \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer my-secret-token' \
  -d '{"value":"bar","ttl_seconds":10}'

# Get
curl -i http://localhost:8080/v1/string/foo \
  -H 'Authorization: Bearer my-secret-token'

# Delete
curl -i -X DELETE http://localhost:8080/v1/string/foo \
  -H 'Authorization: Bearer my-secret-token'

# LPush
curl -i -X POST http://localhost:8080/v1/list/mylist/push \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer my-secret-token' \
  -d '{"items":["a","b","c"]}'

# RPop
curl -i -X POST http://localhost:8080/v1/list/mylist/pop \
  -H 'Authorization: Bearer my-secret-token'
```

---

## Testing

```bash
go test ./...
```

---

## Notes & Extensions

- **List ops**: only `LPush`/`RPop` are implemented for demonstration purposes; could add `RPush`, `LPop`, `LLen`, `LRange`, etc.
- **Context propagation**: all public methods accept `context.Context` to future-proof for I/O, tracing, and cancellation.
- **Storage backends**: we can easily swap in Redis, PostgreSQL, MySQL, etc., by implementing `domain.EntryRepository`.
- **Metrics & Tracing**: integrate Prometheus, OpenTelemetry, etc., via middleware.
- **Set vs Update**:  
  `SetString` currently performs an “upsert” (it creates a new key or overwrites an existing one).  
  If a true “update-only” operation (error if the key doesn’t exist) is required, we could add an `UpdateString` method in the service layer and expose it via a `PATCH /v1/string/{key}` endpoint and a `--action=update` CLI command.