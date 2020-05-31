# prediction-league

## Requirements

Either...

(CPU-intensive)

* Docker Compose
* Make

Or...

(Performance stability)

* Golang 1.14
* npm 13.10 (or nvm)
* Docker Compose
* Make

## Getting Started

### Local Environment

This project can be run either fully-Dockerised (runs exclusively within containers) or
partly-Dockerised (dependencies and build tools are in containers, but the service runs
using Golang on the host).

At the very least, this project runs a MySQL container for a persistent data store, as well as an
"assets" (Node) container to take care of all front-end builds. The builds comprise Sass for CSS and
Babel via Webpack for JavaScript (including Vue).

The idea was originally to use a single process (Webpack) to generate both JavaScript and CSS, but
I couldn't get the Sass build process to produce anything other than standard JS output(?), which
generally isn't valid CSS.

I then reverted to running separate Sass and Babel CLIs via npm commands which worked fine until I
added Vue, then suddenly Babel decided it was going to go maverick and not properly transpile a
"require" function or something.

Hence why I had to revert again and we now have a weird Sass CLI/Webpack mashup combo.

Because it sometimes does what it's supposed to.

I prefer Backend...

Anyway, both build processes will watch for changes to CSS/JS/Vue files - however, changing Go files or
HTML files/templates will require the service itself to be restarted each time.

### Run App

#### Fully-Dockerised

Look out for CPU with this option. Where the asset builds are watching for changes across a network, it
can get quite resource-hungry...

```bash
make app.up
```

#### Partly-Dockerised

Requires more dependencies but results in generally quieter fans...

```bash
nvm install 13.10
nvm use 13.10
make app.run
```

#### Stop local environment

```bash
make app.stop
```

#### Restart app container in local environment

```bash
make app.restart
```

#### Destroy local environment

```bash
make app.kill
```

### Run Testsuite

#### Fully-Dockerised

```bash
make test.run
```

#### Partly-Dockerised

```bash
make test.start
go test -v ./...
```

#### Stop local test environment

```bash
make test.stop
```

## Domain Knowledge

### Entities

#### Season

A Season represents a real-world tournament (such as "Premier League 1992/93"). The Seasons data used in the program
is defined in code as a single data structure (see `datastore.Seasons`).

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

Include a new `Season` struct in the map provided by `datastore.Seasons`.

This struct must adhere to the validation rules found within `domain.ValidateSeason()`.

Running the testsuite again will apply the rules to each `Season` in the map and fail if any aren't met.
