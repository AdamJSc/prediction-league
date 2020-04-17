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

## Background

### Season

A season represents a real-world tournament (such as "Premier League 1992/93"), which can only be created by an "admin".

This is intended as a one-off action once a year, immediately prior to accepting entries into a game.

For this reason, full CRUD has not been enabled for Seasons - only a create endpoint.
