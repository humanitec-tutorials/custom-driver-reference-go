# Go Service Template

## How to create a service from this template:
1. Click "Use this template" green button.
2. Select owner and type a name for your repository, optionally set other properties and click "Create repository from template".
3. Clone the new repository.
4. Replace `go-service-template` to your service name everywhere in the repository.
5. Follow `// TODO:` comments to update the code.
6. Update this README.

## Running the server

```go run ./cmd/server [-c <config>] [-h <host>] [-p <port>]```

## Running tests

```
go generate ./...
go test ./... [-v] [-cover] [-race]```
```

## Running tests in docker-compose

```
# Run docker-compose:
docker-compose up --build

# Run tests:
docker-compose exec server go test ./...
```

## Integration tests

Build the image and run it with docker-compose (in the root of the repo):

```
docker-compose up --build
```

## Update Open API spec

```gogenoas -i ./cmd/server/main.go -o ./openapi/spec.json```

NOTE: **gogenoas** tool can be downloaded from https://github.com/Humanitec/gogenoas.

## Configuration

Configuration sources:
* **Environment variables** (override default configuration from YAML file)
* **Command-line arguments** (override all other settings; use `--help` switch to see available commands and options)

| Variable | Default | Description |
|---|---|---|
| `HOST` | `''` | A host name to filter informing requests (`''` = accpet all). |
| `PORT` | `8080` | A port to listen for the incoming requests on. |
| `DATABASE_HOST` | | Database instance host name. |
| `DATABASE_PORT` | | Database instance port. |
| `DATABASE_USER` | | Database instance login (user name). |
| `DATABASE_PASSWORD` | | Database instance password. |
| `DATABASE_NAME` | | Database name. |
| `DD_ENABLE` | `false` | Enables/Disables DataGog monitoring instrumentation. |
| `DD_ENV` | | DataDog environment name: `dev`, `staging`, `prod`, etc. Only used if `DD_ENABLE` is set to `true`. |
| `DD_SERVICE` | | DataDog service display name. Only used if `DD_ENABLE` is set to `true`. |

## Endpoints

### Public

| Method | Path Template | Description |
| --- | --- | ---|
| `GET` | `/entities/{id}` | Returns the Entity details. |

### Internal

| Method | Path Template | Description |
| --- | --- | ---|
| `GET` | `/internal/entities/{id}` | Returns the Entity details. |

### Service

| Method | Path Template | Description |
| --- | --- | ---|
| `GET` | `/docs/spec.json` | OpenAPI v3 specification. |
| `GET` | `/alive` | Should be used for liveness probe. |
| `GET` | `/health` | Should be used for readiness probe. |