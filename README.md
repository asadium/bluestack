# Bluestack

Bluestack is a LocalStack-style emulator for Azure services. It provides minimal but realistic Azure API implementations for local development and testing.

## Overview

Bluestack exposes a single edge HTTP port that receives all requests and routes them to appropriate service modules. Currently, it supports:

- **Azure Blob Storage** - Basic blob operations (create container, upload, download, list, delete)

Future services planned:
- Azure Queue Storage
- Azure Key Vault
- Azure Table Storage
- And more...

## Architecture

### Edge Router

The edge router (`internal/httpx/router.go`) is the single entry point for all HTTP requests. It:
- Listens on `EDGE_PORT` (default: 4566)
- Routes requests to service modules based on path prefixes
- Provides middleware for logging, request ID, recovery, and timeouts
- Includes a health check endpoint at `/health`

### Services

Services implement the `core.Service` interface and register themselves with the service registry. Each service:
- Implements `Name()` to return its identifier
- Implements `RegisterRoutes()` to set up HTTP routes
- Manages its own state through a store interface

### Storage Layer

Services use store interfaces (e.g., `BlobStore`) that can be implemented with different backends:
- **File-based storage** - Stores data as files under `DATA_DIR` (default: `./data`)
- Future: SQLite-based storage for better performance and querying
- Future: In-memory storage for testing

## Getting Started

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Running Locally

1. Clone the repository:
```bash
git clone https://github.com/asad/bluestack.git
cd bluestack
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run ./cmd/bluestack start
```

Or use the Makefile:
```bash
make run
```

The server will start on port 4566 (or the port specified in `EDGE_PORT` environment variable).

### Configuration

Configuration is loaded from environment variables. See `.env.example` for available options:

- `EDGE_PORT` - HTTP port for the edge router (default: 4566)
- `DATA_DIR` - Base directory for service data (default: `./data`)
- `ENABLED_SERVICES` - Comma-separated list of services to enable (default: `blob`)
- `LOG_LEVEL` - Logging level: debug, info, warn, error (default: `info`)

You can create a `.env` file or export these variables:
```bash
export EDGE_PORT=4566
export DATA_DIR=./data
export ENABLED_SERVICES=blob
export LOG_LEVEL=info
```

### Building

Build the binary:
```bash
make build
```

Or manually:
```bash
go build -o bin/bluestack ./cmd/bluestack
```

### Docker

Build the Docker image:
```bash
make docker-build
```

Or manually:
```bash
docker build -t bluestack:latest -f docker/Dockerfile .
```

Run the container:
```bash
make docker-run
```

Or manually:
```bash
docker run -p 4566:4566 -v $(pwd)/data:/app/data bluestack:latest
```

## Usage Examples

### Health Check

```bash
curl http://localhost:4566/health
```

Response:
```json
{"status":"healthy","service":"bluestack"}
```

### Blob Storage Operations

#### Create a Container

```bash
curl -X PUT http://localhost:4566/blob/myaccount/mycontainer
```

#### Upload a Blob

```bash
echo "Hello, Bluestack!" | curl -X PUT \
  -H "Content-Type: text/plain" \
  --data-binary @- \
  http://localhost:4566/blob/myaccount/mycontainer/myblob.txt
```

#### Download a Blob

```bash
curl http://localhost:4566/blob/myaccount/mycontainer/myblob.txt
```

#### List Blobs

```bash
curl "http://localhost:4566/blob/myaccount/mycontainer?list"
```

#### Delete a Blob

```bash
curl -X DELETE http://localhost:4566/blob/myaccount/mycontainer/myblob.txt
```

#### Delete a Container

```bash
curl -X DELETE http://localhost:4566/blob/myaccount/mycontainer
```

## Testing

Run tests:
```bash
make test
```

Or manually:
```bash
go test ./...
```

## Project Structure

```
bluestack/
├── cmd/
│   └── bluestack/
│       └── main.go              # Application entry point
├── internal/
│   ├── cli/
│   │   └── root.go              # CLI commands (cobra)
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── core/
│   │   └── services.go          # Service interface and registry
│   ├── httpx/
│   │   └── router.go            # Edge HTTP router
│   ├── logging/
│   │   └── logger.go            # Structured logging (zap)
│   ├── services/
│   │   └── blob/
│   │       ├── blob_service.go  # Blob service HTTP handlers
│   │       ├── blob_store.go    # Blob storage implementation
│   │       ├── models.go        # Blob data models
│   │       └── blob_service_test.go  # Tests
│   └── state/
│       └── state.go             # State management (placeholder)
├── docker/
│   └── Dockerfile               # Multi-stage Docker build
├── .env.example                 # Example environment variables
├── Makefile                     # Build and run targets
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## Extending Bluestack

### Adding a New Service

1. Create a new package under `internal/services/<service-name>/`
2. Implement the `core.Service` interface:
   ```go
   type MyService struct {
       // service fields
   }
   
   func (s *MyService) Name() string {
       return "myservice"
   }
   
   func (s *MyService) RegisterRoutes(router chi.Router) {
       // register HTTP routes
   }
   ```
3. Register the service in `internal/cli/root.go`:
   ```go
   myService := myservice.NewMyService(store, logger)
   core.RegisterService(myService)
   ```
4. Add the service name to `ENABLED_SERVICES` environment variable

### Adding a New Storage Backend

1. Implement the service's store interface (e.g., `BlobStore`)
2. Update the service initialization to use the new store
3. Consider adding configuration options to choose between backends

## Limitations

- **Not Azure-compliant**: This is a simplified emulator, not a full Azure implementation
- **No authentication**: Currently no authentication/authorization (add as needed)
- **Limited features**: Only basic operations are implemented
- **Single process**: Not designed for distributed deployment

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see LICENSE file for details.

