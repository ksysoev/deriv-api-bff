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
- **Passthrough for Deriv API**: All request that are not following format of BFF will be forwared to Deriv API as is

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

## Commands

### Server

Start the BFF server with the specified configuration file:

```sh
bff server --config=./config.yaml
```

### Config Verify

Verify the correctness of the API call configuration:

```sh
bff config verify --config=./config.yaml
```

### Config Upload

Upload the API call configuration to the remote source (e.g., etcd):

```sh
bff config upload --config=./config.yaml
```

## Command Line Arguments

Each command supports the following arguments:

- `--config string`: Specifies the path to the configuration file that the BFF service will use.
- `--loglevel string`: Sets the logging level for the application. Options include "debug", "info", "warn", and "error". The default level is "info".
- `--logtext`: When set, logs will be output in plain text format. If not set, logs will be in JSON format.
- `--api-source-path string`: Defines the path to the API configuration file or directory. The BFF service will read and merge all YAML files in the specified directory.
- `--api-source-etcd-servers string`: Provides a comma-separated list of etcd servers to be used as the source for API configuration.
- `--api-source-etcd-prefix string`: Specifies the prefix for API configuration keys in etcd.

**Example usage**:

```sh
bff server --config=./config.yaml --loglevel=debug --logtext
```



## Server Configuration

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

## Environment Variables

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

## API Configuration

The API configuration can be specified in a YAML file or a directory containing multiple YAML files. If a directory is provided, the BFF service will scan all YAML files in that directory and merge them into a single configuration.

### Format of API Call Definitions

Each API call should be defined in the following format:

- `method`: The name of the API call.
- `params`: JSON schema definition for all parameters.
- `backend`: A list of definitions for upstream API calls.

Backends can have two types of upstream requests:

1. **Deriv API Request**
2. **HTTP API Request**

### Deriv API Request

To define a Deriv API request, use the following parameters:

- `name`: (Optional) Name for the API call. If you want to depend on this API call and use its data in other API calls, you need to provide a name.
- `request`: Request object that represents the template of the future request.
- `allow`: Fields that will be copied to the final response. If the response is an object, the fields will be copied directly. If the response is an array, the BFF will create an object with `list` as the key and the response as the value. If the response is a scalar value, the key `value` will be used.
- `fields_map`: Allows renaming fields in the final response.
- `depends_on`: Defines dependencies on other API calls. If dependencies are defined, their response data can be used in the request template.

### HTTP API Request

To define an HTTP API request, use the following parameters:

- `name`: (Optional) Name for the API call. If you want to depend on this API call and use its data in other API calls, you need to provide a name.
- `method`: HTTP method for the request.
- `url`: Template for the URL.
- `headers`: Templates for HTTP headers.
- `request`: Template for the body of the HTTP request.
- `allow`: Fields that will be copied to the final response. If the response is an object, the fields will be copied directly. If the response is an array, the BFF will create an object with `list` as the key and the response as the value. If the response is a scalar value, the key `value` will be used.
- `fields_map`: Allows renaming fields in the final response.
- `depends_on`: Defines dependencies on other API calls. If dependencies are defined, their response data can be used in the request template.


### Template Placeholders

Template placeholders are supported in the values of `request` and `url` for HTTP requests. Placeholders should follow the format `${path.to.the.key}`.

The following data is provided to each template:

- `params`: Object with incoming parameters defined in the `params` section.
- `resp`: If the API call has defined dependencies, all responses will be provided as part of this object. You can use the name of the dependency to reference fields from it.
- `req_id`: ID of the API request, which can be used for tracing.

### Example Configuration

```yaml
- method: "aggregate" 
  params:
    country: 
      type: "string"
  backend:
    - allow: 
        - virtual_company
        - name
      fields_map:
        financial_company: "financial"
        gaming_company: "gaming"
      request: 
        landing_company: "${params.country}"
    - allow: 
        - api_call_limits
        - clients_country
      request:
        website_status: 1
- method: "chain"
  backend:
    - depends_on: 
        - "website_status"
      allow: 
        - virtual_company
        - name
      fields_map:
        financial_company: "financial"
        gaming_company: "gaming"
      request:
        landing_company: '${resp.website_status.clients_country}'
    - name: "website_status"
      allow: 
        - api_call_limits
        - clients_country
      request: 
        website_status: 1
- method: "config_country"
  params:
    country: 
      type: "string"
    token:
      type: "string"
  backend:
    - method: "GET"
      headers:
        Authorization: "Bearer ${params.token}"
      url: 'http://nginx/country/${params.country}.json'
      allow: 
        - region
        - subregion
        - name
        - alpha2Code
        - alpha3Code
        - callingCodes
        - capital
      fields_map:
        title: "title"
```

## Request Format

When making API requests to the BFF service, the request should be structured as follows:

- `method` (string, required): The name of the API call.
- `params` (object, optional): An object containing the parameters for the API call. Only fields defined in the `params` configuration for the API call are accepted.
- `req_id` (string, optional): A unique identifier for the request. This can be used to match the response with the request on the client side.
- `passthrough` (object, optional): An object for passing additional context. This object will be included in the response without modification.

### Example API Request

Here is an example of a properly formatted API request:

```json
{
    "method": "config_country",
    "params": {
        "country": "id"
    },
    "req_id": "123",
    "passthrough": {
        "key": 1
    }
}
```

### Explanation

- **method**: Specifies the API call to be executed. This field is mandatory.
- **params**: Contains the parameters required for the API call. This field is optional and should match the schema defined in the API configuration.
- **req_id**: A unique identifier for the request. This field is optional and helps in tracking the request and response.
- **passthrough**: Allows additional context to be passed through the request. This field is optional and will be included in the response as-is.

### Notes

- Ensure that the `method` field matches one of the API calls defined in your configuration.
- The `params` object should only include fields that are defined in the `params` section of the corresponding API call configuration.
- The `req_id` can be any number that uniquely identifies the request. It is useful for debugging and tracking purposes.
- The `passthrough` object can contain any additional data you want to include in the response without modification.

By following this format, you can ensure that your API requests are correctly structured and processed by the BFF service.

## Response Format

The response format follows Deriv's API structure and includes the following fields:

- **echo**: Contains the original request.
- **msg_type**: Indicates the type of message. For a successful response, it will contain the method name. For an error response, it will contain `error`.
- **response body**: The actual response data will be placed in a field named after the `msg_type`.
- **req_id** (optional): This field will be included in the response if a request ID was provided in the request.
- **passthrough** (optional): This field will be included in the response if a passthrough object was provided in the request.

### Example Response

Here is an example of a properly formatted response:

```json
{
    "echo": {
        "method": "config_country",
        "params": {
            "country": "id"
        },
        "req_id": "123",
        "passthrough": {
            "key": 1
        }
    },
    "msg_type": "config_country",
    "config_country": {
        "region": "Asia",
        "subregion": "Southeast Asia",
        "name": "Indonesia",
        "alpha2Code": "ID",
        "alpha3Code": "IDN",
        "callingCodes": ["62"],
        "capital": "Jakarta"
    },
    "req_id": "123",
    "passthrough": {
        "key": 1
    }
}
```

### Explanation

- **echo**: Reflects the original request, including the method, params, req_id, and passthrough fields.
- **msg_type**: Indicates the type of message. In this example, it is `config_country`, which matches the method name in the request.
- **response body**: Contains the actual response data. In this example, the response data is placed in a field named `config_country`.
- **req_id**: Matches the request ID provided in the request, allowing for easy correlation between requests and responses.
- **passthrough**: Includes any additional context provided in the request, passed through unchanged.

### Notes

- The **msg_type** field helps identify the type of response and matches the method name for successful responses.
- The **response body** field is dynamically named based on the **msg_type**.
- The **req_id** and **passthrough** fields are optional and will only be included if they were part of the original request.

By following this format, you can ensure that your API responses are correctly structured and consistent with Deriv's API standards.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue if you encounter any problems or have suggestions for improvements.

## License

This project is licensed under the MIT License.
