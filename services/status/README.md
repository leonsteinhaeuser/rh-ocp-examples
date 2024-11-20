# status service

This service takes a comma separated list of key=value pairs and returns the status of the services in the list via the `/status` endpoint.

## Configuration

| Environment variable | Type | Default | Description |
| -------------------- | ---- | ------- | ----------- |
| `LISTEN_ADDRESS` | string | `:8082` | Address to listen on for HTTP requests. |
| `EXTERNAL_SERVICES_TO_WATCH` | []string | `redhat=https://www.redhat.com/en` | Comma separated list of key=value pairs of services to watch |
| `CHECK_INTERVAL` | duration | `5s` | Interval parsed as duration, to check the upstream service. |
| `REQUEST_TIMEOUT` | duration | `2s` | Timeout parsed as duration, for the HTTP request to the upstream service. |

## API

| Path | Method | Description |
| ---- | ------ | ----------- |
| `/status` | GET | Returns the status of the services in the list. |
| `/healthz` | GET | Returns the health of the service. |
