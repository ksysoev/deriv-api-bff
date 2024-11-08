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


## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue if you encounter any problems or have suggestions for improvements.


## License

This project is licensed under the MIT License.
