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
* or using native Golang: `go run service/main.go`

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
* or using native Golang: `go test -v ./...`

To stop the dependencies:

```bash
make test.stop
```

## Domain Knowledge

### Entities

#### Season

A Season represents a real-world tournament (such as "Premier League 1992/93"). The Seasons data used in the program
is defined in code as a single data structure (see `domain.Seasons()`).

This is therefore deliberately controlled by the project maintainer as a one-off action approximately once a year,
immediately prior to accepting entries into a game.

#### Team

#### Entry

#### Realm

This is effectively an arbitrary flag that represents a distinct instance of the "game".

Each entry pertains to a particular Realm which effectively acts as a sub-grouping of Entries. So two Entries that have
different Realm values belong to different "games" and are therefore not competing against each other.

By default, and as a "quick-win", the Realm is determined by the domain/host name via which an Entry is created.

This means that the program can run as a single process serving multiple domains/sub-domains, with each one operating as
its own independent instance of the game.

Each Realm should have a corresponding PIN, which must be supplied when creating an Entry. This is configured within
the program's corresponding `.env` file, using the key name that is in the format `<RealmName>_REALM_PIN`.

e.g. To set the PIN pertaining to `my.domain.com`, use the Env Key `MY_DOMAIN_COM_REALM_PIN`.
To set the PIN pertaining to your local instance, use the Env Key `LOCALHOST_REALM_PIN`.

Each Realm should also be associated with the ID of a Season that is currently active for that Realm. This enables multiple
different Realms to be playing under multiple Seasons at the same time, so one Realm could be assigned `199293_1`
("Premier League 1992/93") whilst at the same time another Realm is assigned  `199293_2` ("Division One 1992/93").

This is configured within the program's corresponding `.env` file, using the key name that is in the format
`<RealmName>_REALM_SEASON_ID`.

e.g. To set the Season ID pertaining to `my.domain.com`, use the Env Key `MY_DOMAIN_COM_REALM_SEASON_ID`.
To set the Season ID pertaining to your local instance, use the Env Key `LOCALHOST_REALM_SEASON_ID`.

The values of each `*_REALM_PIN` and `*_REALM_SEASON_ID` key are consolidated as `Realm` objects and inflated on a `domain.Config`
object within the main bootstrap, so that these can be passed around within the centralised object as required.

#### Guard

A Guard is an arbitrary means of assessing whether an agent method should continue its execution, and can be thought of
as a very basic and rudimentary per-method permission system.

There are two component parts to a Guard: the `Attempt` and the `Target`.

The `Attempt` is usually the value supplied within a request in some way.

The `Target` is usually a "known" value that the `Attempt` must match in order to return a positive result to the agent method.

For example...

Most route handlers will at some point invoke `ctx := domain.NewContextFromRequest(r, config)` to generate a `domain.Context` object,
which comprises an arbitrary `Guard` field, as well as a `Realm` field that has been populated with details of the system's
current Realm.

The route handler might then invoke `ctx.Guard.SetAttempt(input.PIN)` to set the Guard's Attempt value - i.e. the value
"attempting" to match the Target.

(In this case, our Guard Attempt is the "PIN" field of the incoming request body, so our agent method will want this to
match the PIN of the system's current Realm in order that it can allow the operation to continue).

The handler will pass this context to the main agent method it invokes (e.g. `domain.EntryAgent.CreateEntry(ctx ....)`).

The agent method can then invoke `ctx.Guard.AttemptMatchesTarget(ctx.Realm.PIN)` to determine if the incoming request
can be authorised.

## Maintenance

### To add a new Season...

Include a new `Season` struct in the map returned by `domain.Seasons()`.

This struct must adhere to the validation rules found within `domain.ValidateSeason()`.

Running the testsuite again will apply the rules to each `Season` in the map and fail if any aren't met.
