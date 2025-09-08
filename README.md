# dis-bundle-api

The dis-bundle-api is a backend service for managing and publishing datasets and content as bundles, similar to Florence’s collections.

## Getting started

* Run `make debug` to run application on http://localhost:29800
* Run `make help` to see full list of make targets

To run `make lint-api-spec` you require the Node version specified in [`.nvmrc`](.nvmrc) and to install `redocly/cli`:

```sh
   # Ensure you have the correct version of node installed
   nvm install

   # Install redocly
   nvm exec -- npm install -g @redocly/cli
```

### Dependencies

* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable              | Default                  | Description                                                                                                        |
| --------------------------------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------ |
| BIND_ADDR                         | `:29800`                 | The host and port to bind to                                                                                       |
| DATASET_API_URL                   | `http://localhost:22000` | The hostname and port for the Dataset API                                                                          |
| GRACEFUL_SHUTDOWN_TIMEOUT         | `5s`                     | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_INTERVAL              | `30s`                    | Time between self-healthchecks (`time.Duration` format)                                                            |
| HEALTHCHECK_CRITICAL_TIMEOUT      | `90s`                    | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| OTEL_BATCH_TIMEOUT                | `5s`                     | Timeout for OpenTelemetry batch export (`time.Duration` format)                                                    |
| OTEL_EXPORTER_OTLP_ENDPOINT       | `localhost:4317`         | Endpoint for OpenTelemetry service                                                                                 |
| OTEL_SERVICE_NAME                 | `dis-bundle-api`         | Label of service for OpenTelemetry service                                                                         |
| OTEL_ENABLED                      | `false`                  | Feature flag to enable OpenTelemetry                                                                               |
| DEFAULT_MAXIMUM_LIMIT             | `1000`                   | Default number of maximum bundles returned                                                                         |
| DEFAULT_LIMIT                     | `20`                     | Default number of bundles returned                                                                                 |
| DEFAULT_OFFSET                    | `0`                      | Default offset                                                                                                     |
| ENABLE_PERMISSIONS_AUTH           | `false`                  | Feature flag to enable permissions authentication                                                                  |
| SLACK_ENABLED                     | `false`                  | Feature flag to enable Slack notifications                                                                         |
| ZEBEDEE_URL                       | `http://localhost:8082`  | Zebedee URL                                                                                                        |
| ZEBEDEE_CLIENT_TIMEOUT            | `30s`                    | Timeout for Zebedee client (`time.Duration` format)                                                                |

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2025, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
