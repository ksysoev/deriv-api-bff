# Deriv BFF Service

[![Deriv BFF CI](https://github.com/ksysoev/deriv-api-bff/actions/workflows/main.yml/badge.svg)](https://github.com/ksysoev/deriv-api-bff/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/ksysoev/deriv-api-bff/graph/badge.svg?token=2YOCWTOBV7)](https://codecov.io/gh/ksysoev/deriv-api-bff)

This project implements the Backend for Frontend (BFF) pattern on top of the Deriv public API. The main goal is to provide a dedicated backend for frontend applications, optimizing the interaction between the client and the server.

## Features

- **Aggregate Information**: Combine data from multiple Deriv API calls into a single response.
- **Filter Response Data**: Send only the desired fields to the client.
- **Multi-Step Sequences**: Build complex workflows with multiple API requests.
- **Extend Deriv API**: Integrate your own HTTP APIs seamlessly.
- **Declarative API Creation**: Create new API calls in a declarative way without writing code.

## Installation

### From Source Code

To install from the source code, run the following command:

```sh 
go install https://github.com/ksysoev/deriv-api-bff/cmd/bff@latest
```

### With Docker Image

To pull the Docker image, run the following command:

```sh 
docker pull ghcr.io/ksysoev/deriv-api-bff:latest
```

### Commands

#### Server

Start the BFF server with the specified configuration file:

```sh
bff server --config=./config.yaml
```

#### Config Verify

Verify the correctness of the API call configuration:

```sh
bff config verify --config=./config.yaml
```

#### Config Upload

Upload the API call configuration to the remote source (e.g., etcd):

```sh
bff config upload --config=./config.yaml
```

### Command Line Arguments

Each command supports the following arguments:

- `--config string`: Path to the configuration file (default: "./runtime/config.yaml").
- `--loglevel string`: Log level (options: "debug", "info", "warn", "error") (default: "info").
- `--logtext`: Log in text format; if not set, logs will be in JSON format.

Example usage:

```sh
bff server --config=./config.yaml --loglevel=debug --logtext
```

### Server Configuration

The server configuration can be specified in a YAML file. Below is an example configuration:

```yaml
server:
  listen: ":8080"  # The address and port on which the server listens
  max_requests: 100  # Maximum number of concurrent requests the server can handle
  max_requests_per_conn: 10  # Maximum number of concurrent requests per client connection

deriv:
  endpoint: "wss://ws.derivws.com/websockets/v3"  # Deriv API endpoint

otel:
  prometheus:
    listen: ":8081"  # The address and port for Prometheus metrics
    path: "/metrics"  # The path for Prometheus metrics

api_source:
  etcd:
    servers: "etcd:2379"  # Etcd server address
    prefix: "api::"  # Prefix for API configuration keys in Etcd
  path: "./runtime/api_config"  # Path to the local API configuration directory
```

### Environment Variables

You can also configure the server using environment variables. Below are the available environment variables:

```sh
SERVER_LISTEN=:8080  # The address and port on which the server listens
SERVER_MAX_REQUESTS=100  # Maximum number of concurrent requests the server can handle
SERVER_MAX_REQUESTS_PER_CONN=10  # Maximum number of concurrent requests per client connection
DERIV_ENDPOINT=wss://ws.derivws.com/websockets/v3  # Deriv API endpoint
OTEL_PROMETHEUS_LISTEN=:8081  # The address and port for Prometheus metrics
OTEL_PROMETHEUS_PATH=/metrics  # The path for Prometheus metrics
API_SOURCE_ETCD_SERVERS=etcd:2379  # Etcd server address
API_SOURCE_ETCD_PREFIX=api::  # Prefix for API configuration keys in Etcd
API_SOURCE_PATH=./runtime/api_config  # Path to the local API configuration directory
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue if you encounter any problems or have suggestions for improvements.


## License

This project is licensed under the MIT License.
