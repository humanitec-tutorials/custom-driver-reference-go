# Humanitec Custom Resource Driver - Reference implementation

This is a **reference** implementation of a custom *Driver* for Humanitec.
It operates similarly to the Built-in AWS driver, deploying new S3 buckets on-demand.

To test it out, you'll need to deploy it as a public facing webserver, and then register it via the Humanitec API.
To learn more about registering drivers checkout the [documentation](https://docs.humanitec.com/integrations/create-own-resource-driver).


## Endpoints

### Public

| Method   | Path Template | Description                |
|----------|---------------|----------------------------|
| `PUT`    | `/s3/{id}`    | Upsert a bucket on AWS S3. |
| `DELETE` | `/s3/{id}`    | Delete a bucket on AWS S3. |


### Service

| Method | Path Template     | Description                         |
|--------|-------------------|-------------------------------------|
| `GET`  | `/docs/spec.json` | OpenAPI v3 specification.           |
| `GET`  | `/alive`          | Should be used for liveness probe.  |
| `GET`  | `/health`         | Should be used for readiness probe. |


## Configuration

Configuration sources:
* **Environment variables** (override default configuration from YAML file)
* **Command-line arguments** (override all other settings; use `--help` switch to see available commands and options)

| Flag | Variable          | Default  | Description                                                            |
|------|-------------------|----------|------------------------------------------------------------------------|
| -h   | `HOST`            | `''`     | The ip to listen for incoming requests on (`''` = accept all).         |
| -p   | `PORT`            | `8080`   | The port to listen for the incoming requests on (default is 8080).     |
| -l   | `LOG_LEVEL`       | `'info'` | The level of logging expected (`'info'`,`'warn'`,`'error'`,`'debug'`). |
| -m   | `FAKE_AWS_CLIENT` | `false`  | Use the mock AWS API (for unitests).                                   |



## Running the server

```go run ./ [-c <config_file>] [-h <host>] [-p <port>] -```

