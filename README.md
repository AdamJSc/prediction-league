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
using npm/nvm and Golang on the host).

At the very least, this project runs a MySQL container for a persistent data store. The builds comprise Sass for CSS and
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
make app.docker.up
```

#### Partly-Dockerised

Requires more dependencies but results in generally quieter fans...

Make sure you're using npm version 13.10

```bash
nvm install 13.10
nvm use 13.10
```

Then start your engines:

```bash
make app.install
make app.run
```

### Run Testsuite

#### Fully-Dockerised

```bash
make test.docker.up
```

#### Partly-Dockerised

```bash
make test.run
```

Take a look at the other `make` commands in the project root `Makefile` which help to automate some of the stop/restart/kill workflows.

## Entities

### Season

A Season represents a real-world tournament (such as "Premier League 1992/93"). The Seasons data used in the program
is defined in code as a single data structure (see `datastore.Seasons`).

This is therefore deliberately controlled by the project maintainer as a one-off action approximately once a year,
immediately prior to accepting entries into a game.

For details on the system's default season, see ["FakeSeason"](#fakeseason) (below).

### Team

### Entry

### Entry Prediction

### Standings

### Scored Entry Prediction

### Ranking

### Ranking With Meta

### Ranking With Score

### Ranking With Score And Meta Position

### LeaderBoard

### LeaderBoard Ranking

### Token

### Email Message

### Message Identity

## Other Domain Knowledge

### Ingesting real-world standings

### Realm

This is an arbitrary flag that represents a distinct instance of the "game".

Each Entry belongs to a particular Realm which effectively acts as a sub-grouping of Entries. So two Entries that have
different Realm values belong to different "games" and are therefore not competing against each other.

By default, and as a "quick-win", the Realm is determined by the domain/host name via which an Entry is created.

This means that the program can run as a single process serving multiple domains/sub-domains, with each one operating as
its own independent instance of the game.

Realms must be configured in advance, within the `./data/realms.yml` file. This file's payload comprises the single
root-level key named `realms`, and each key beneath this should match the lowercase realm name (domain) -
e.g. `localhost` or `my.sub.domain.com`. See the file itself for an example.

Each Realm should have a corresponding PIN which must be supplied when creating an Entry, in order to prevent unwelcome sign-ups.

This value is taken from the `pin` key of the realm config described above.

Each Realm should also be associated with the ID of a Season that is currently active for that Realm. This enables multiple
different Realms to be playing under multiple Seasons at the same time, so one Realm could be assigned `199293_1`
("Premier League 1992/93") whilst at the same time another Realm is assigned  `199293_2` ("Division One 1992/93").

This value is taken from the `season_id` key of the realm config described above.

The values of these payloads are parsed as `Realm` objects and inflated on a `domain.Config` object within the main
bootstrap, so that these can be passed around within the centralised object as required.

### Guard

A Guard is an arbitrary means of assessing whether an agent method should continue its execution, and can be thought of
as a very basic and rudimentary per-method permission system.

There are two component parts to a Guard: the `Attempt` and the `Target`.

The `Attempt` is usually the value supplied within a request in some way.

The `Target` is usually a "known" value that the `Attempt` must match in order to return a positive result to the agent method.

For example...

Most route handlers will at some point invoke `ctx := contextFromRequest(r, c)`, where `r` is a HTTP Request object,
and `c` is the container comprising all of our app's dependencies. This will return a new context with two values added:
one is an empty `Guard` object, and the other is a `Realm` object that has been populated with details of the request's
current Realm.

The route handler might then invoke `domain.GuardFromContext(ctx).SetAttempt(input.PIN)` to set the Guard's Attempt value -
i.e. the value "attempting" to match the Target - using some data point that has been supplied in the user's request.

(In this case, our Guard Attempt is the "PIN" field of the incoming request body, so our agent method will want this to
match the PIN of the request's current Realm in order that it can allow the operation to continue).

The handler will pass this context to the main agent method it invokes (e.g. `domain.EntryAgent.CreateEntry(ctx ....)`).

The agent method can then invoke `domain.GuardFromContext(ctx).Guard.AttemptMatchesTarget(ctx.Realm.PIN)` which returns
a `boolean` to determine whether or not the incoming request can be authorised.

### "FakeSeason"

Much of the functionality within this system is time-sensitive in relation to the Season that applies to the current
Realm. Functionality such as whether or not an Entry or Entry Prediction can be created or updated at any given moment
is determined by the corresponding timeframes set on the Season itself.

Usually these will be **absolute** timeframes that pertain to the dates relevant to a real-world Season
(see `201920_1` in `datastore.Seasons` as an example).

For this reason, the default Realm (`localhost`) is affiliated with a specific Season that has the ID `FakeSeason`.
This Season's timeframes are **relative** to the point at which the system is run, so that core functionality can be
used immediately.

The schedule takes place as follows:

* 0 mins - 20 mins
    * Entries can be created or updated
    * Entry Predictions can be created
* 20 mins - 40 mins
    * No Entries can be created or updated
    * No Entry Predictions can be created
* 40 mins - 60 mins
    * No Entries can be created or updated
    * Entry Predictions can be created
* 60 mins+
    * No Entries can be created or updated
    * No Entry Predictions can be created

## Cron Tasks

## Maintenance

### To add a new Season...

Include a new `Season` struct in the map provided by `datastore.Seasons`.

This struct must adhere to the validation rules found within `domain.ValidateSeason()`.

Running the testsuite again will apply the rules to each `Season` in the map and fail if any aren't met.

## Features To Build

* TBC
* Additional comms methods - HTML email + SMS
* Additional comms events - prediction window opens, prediction window pre-close reminder

## Improvements

### Guard

### Short Codes / Passwords

* Lines blurred as development has gone on
* Also, expired short codes aren't retained, so could **theoretically** be generated again in the future

### Tokens

* Robustness
* Clean up of expired tokens

### Cookie Management

### Scheduled Tasks Concurrency

* Make processing of Standings concurrent when retrieving Latest Standings + scoring Entry Predictions
* Process runs once every 15 minutes and scale currently very small

### Queue retries

* If sending of email fails, retry it X times with a cool-off in between each

### Datastore

* Hard-coded seasons and teams data
* Replace with a "seasons data provider" and "teams data provider" that can be injected via container

### Vue integration

* Pulling from CDN prevents debug mode in development
* Webpack was warning against huge chunks otherwise

### Split out tests

### Cron

* Move it to a runnable process?
