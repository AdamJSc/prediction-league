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

## Domain Knowledge

### Entities

#### Season

A Season represents a real-world tournament (such as "Premier League 1992/93"). The Seasons data used in the program
is defined in code as a single data structure (see `domain.Seasons()`).

This is therefore deliberately controlled by the project maintainer as a one-off action once a year, immediately
prior to accepting entries into a game.

#### Team

#### Entry

#### Realm

This is effectively an arbitrary flag that represents a distinct instance of the "game".

Each entry pertains to a particular Realm which effectively acts as a sub-grouping of Entries. So two Entries that have
different Realm values do not belong to the "game" and are therefore not competing against each other.

By default, and as a "quick-win", the Realm is determined by the domain/host name via which an Entry is created.

This means that the program can run as a single process serving multiple domains/sub-domains, with each one operating as
its own independent instance of the game.

Each Realm should have a corresponding PIN, which is validated upon creation of an Entry. This is configured within
the program's corresponding `.env` file, using the key name that is in the format `<RealmName>_REALM_PIN`.

e.g. To set the PIN pertaining to `my.domain.com`, use the Env Key `MY_DOMAIN_COM_REALM_PIN`.
To set the PIN pertaining to your local instance, use the Env Key `LOCALHOST_REALM_PIN`.

Each Realm should also be associated with the ID of a Season that is currently active for that Realm. This enables multiple
Realms to be playing under multiple Seasons at the same time, so one Realm could be assigned `199293_1` ("Premier League 1992/93")
whilst at the same time another Realm is assigned  `199293_2` ("Division One 1992/93").

This is configured within the program's corresponding `.env` file, using the key name that is in the format
`<RealmName>_REALM_SEASON_ID`.

e.g. To set the Season ID pertaining to `my.domain.com`, use the Env Key `MY_DOMAIN_COM_REALM_SEASON_ID`.
To set the Season ID pertaining to your local instance, use the Env Key `LOCALHOST_REALM_SEASON_ID`.

## Maintenance

### To add a new Season...

Include a new `Season` struct in the map returned by `domain.Seasons()`.

This struct must adhere to the validation rules found within `domain.ValidateSeason()`.

Running the testsuite again will apply the rules to each `Season` in the map and fail if any aren't met.
