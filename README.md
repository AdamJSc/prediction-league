# prediction-league

## Requirements

* Docker
* Docker Compose
* Make
* (optional) Golang 1.14

## Getting Started

To start up app dependencies:

```bash
make app.start
```

Then run the app:

* via Docker: `make app.run`
* native Golang: `go run service/main.go`

To stop the app and dependencies:

```bash
make app.stop
```

## Testing

To start up test dependencies:

```bash
make test.start
```

Then run the testsuite:

* via Docker: `make test.run`
* native Golang: `go test -v ./...`

To stop the dependencies:

```bash
make test.stop
```
