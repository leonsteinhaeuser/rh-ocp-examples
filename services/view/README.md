# View service

This service talks to the number service and displays the number in a web page.

## API

| Endpoint | Method | Payload Request | Payload Response | Description |
| --- | --- | --- | --- | --- |
| / | GET | - | html | Get a random number and display it in a web page |
| /api/v1/status

## Configuration

The view service requires the following environment variables:

- `NUMBER_SERVICE_URL`: The URL of the number service
- `STATUS_SERVICE_URL`: The URL of the status service (including the endpoint)
