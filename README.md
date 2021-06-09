# Prediction League ⚽️‍ 🔮 🧙‍

**v1.1.4**

## About

### Background

Prediction League is a game that started socially between friends.

Everyone in the group puts some coins into the kitty and writes down what they think the final Premier League table will
look like, before a ball has even been kicked.

Whoever is closest at the end of the season wins the kitty. And sometimes gets to pick the following week’s teams for
6-a-side...

### Who needs pen-and-paper?? 📝

This project is a digital representation of the game, which implements a Frontend and Backend for handling new entries,
payment workflows, management of predictions and scoring of points (which are now accrued on a cumulative basis throughout
the season instead of just at the end).

It's written in [Golang](https://golang.org/), using Go [templates](https://golang.org/pkg/text/template/) for HTML.

[Vuejs](https://vuejs.org/) and [Bootstrap](https://getbootstrap.com/) are used on the Frontend.
[Sass](https://sass-lang.com/) is used for pre-processing CSS and [npm](https://npmjs.com/) + [Webpack](https://webpack.js.org/)
are used for building assets.

It's hosted at [prem.footyga.me](https://prem.footyga.me) (current Premier League season) and
[champ.footyga.me](https://champ.footyga.me) (current EFL Championship season).


## Local Environment ♻️ 🌳

To read about the passive-aggressive joys of trying to scaffold this project in the first place, see [docs/local-environment-setup.md](docs/local-environment-setup.md)

### Configuring Dependencies

The following environment variables must be set in order to configure the usage of third-party dependencies.

You can do this by creating a new `.env` file in the project root, or overriding the contents of `infra/app.env`
and `infra/app.docker.env` as required.

Leaving these blank will result in the default behaviour as described.

* `PAYPAL_CLIENT_ID`
    * Client ID as required by PayPal's [Basic Checkout Integration](https://developer.paypal.com/docs/checkout/integrate/)
    workflow.
    * If left blank, provides a "skip payment" step during sign-up (for debugging purposes only).

* `MAILGUN_API_KEY`
    * API Key required by [Mailgun](https://www.mailgun.com/) integration for transactional emails.
    * If left blank, dumps content of email to the log without sending.

* `FOOTBALLDATA_API_TOKEN`
    * API Key required by [football-data.org](https://www.football-data.org/) integration for consumption of real-world
    football league table data.
    * If left blank, no latest league table standings are retrieved and processed by the cron job.

### Fully-Dockerised Setup

Look out for CPU with this option. Where the asset builds are watching for changes across a network, it
can get quite resource-hungry...

#### Requirements

* [Docker](https://docker.com/) + [Docker Compose](https://github.com/docker/compose)
* [GNU Make](https://gnu.org/software/make/manual/make.html)

#### Run App

```bash
make app.docker.up
```

#### Run Testsuite

```bash
make test.docker.up
```

## Partly-Dockerised Setup

Requires more dependencies installed on the host machine, but results in generally quieter fans...

### Requirements

* [Golang](https://golang.org/) 1.16
* [npm](https://npmjs.com/) 13.10 (or [nvm](http://nvm.sh))
* [Docker](https://docker.com/) + [Docker Compose](https://github.com/docker/compose)
* [GNU Make](https://gnu.org/software/make/manual/make.html)

### Run App

Make sure you're using npm version 13.10

```bash
# via nvm
nvm install 13.10
nvm use 13.10
```

Then start your engines:

```bash
make app.install
make app.run
```

### Run Testsuite

```bash
make test.run
```

N.B. Take a look at the other `make` commands in the project root `Makefile` which help to automate
some of the stop/restart/kill workflows too.


## Key Concepts

### Entries and Predictions

Users can sign-up to create an [Entry](docs/domain-knowledge.md#entry) and make their first [Prediction](docs/domain-knowledge.md#entryprediction)
before the [Season](docs/domain-knowledge.md#season) begins.

They can make subsequent changes to their [Prediction](docs/domain-knowledge.md#entryprediction) (i.e. create a new one) during
pre-defined windows throughout the [Season](docs/domain-knowledge.md#season).

These timeframes are configured independently for each [Season](docs/domain-knowledge.md#season).

For example, no more [Entries](docs/domain-knowledge.md#entry) can be made once the [Season's](docs/domain-knowledge.md#season) `EntriesAccepted`
timeframe has elapsed.

Additional settings can also be configured for each [Realm](docs/domain-knowledge.md#realm) (an instance of the game which
runs on a particular URL/sub-domain).

### Payments

Each [Entry](docs/domain-knowledge.md#entry) requires a payment in order to be accepted into the game.

This is to fund the kitty that the winner receives at the end of the game.

For convenience and user peace-of-mind, payment is made via PayPal using their [Basic Checkout Integration](https://developer.paypal.com/docs/checkout/integrate/).

Given that the payment flow exists entirely on the Frontend, each [Entry](docs/domain-knowledge.md#entry) must be "approved" by an
Admin in order that payment can be verified manually.

This is a single API endpoint whose Basic Auth credentials can be configured via the `.env` variable named `ADMIN_BASIC_AUTH`.

Payment can be skipped when running locally for debugging purposes, by leaving the `.env` variable named `PAYPAL_CLIENT_ID`
with an empty value.

### Scoring

For every [Game Week](#game-weeks), each [Prediction](docs/domain-knowledge.md#entryprediction) receives a score, which produces
a [Scored Prediction](docs/domain-knowledge.md#scoredentryprediction) for that Game Week.

[Scored Predictions](docs/domain-knowledge.md#scoredentryprediction) receive 1 "penalty point" for each position that the
Prediction has incorrectly placed a [Team](docs/domain-knowledge.md#team).

For example, if Team A are in 3rd place but are predicted to finish 1st, `2` penalty points will be received.

Likewise, placing Team B at the bottom (in 20th), who currently sit in 15th, will cause `5` penalty points to be issued.

Placing a team in the exact position they are in the real-world table will receive `0` penalty points.

### Game Weeks

Each [Season](docs/domain-knowledge.md#season) is broken down into a number of "Game Weeks", which work exactly the same as in
[Fantasy Football](https://fantasy.premierleague.com/).

i.e. For a Premier League season with 20 teams, there are 38 Game Weeks.
For a Championship season with 24 teams, there are 46 Game Weeks.

Once a new Game Week starts, the previous Game Week’s score is frozen and the next Game Week begins with a score of `0`.

### LeaderBoard

The total cumulative score for all [Scored Predictions](docs/domain-knowledge.md#scoredentryprediction) is calculated and used to
produce the [LeaderBoard](docs/domain-knowledge.md#leaderboard). The [LeaderBoard](docs/domain-knowledge.md#leaderboard)
is topped by the [Entry](docs/domain-knowledge.md#entry) with the lowest cumulative score of penalty points.

In the event of a tie on the [LeaderBoard](docs/domain-knowledge.md#leaderboard), the tied [Entries](docs/domain-knowledge.md#entry) will
be ordered by the **minimum** score that each one has gained throughout all Game Weeks so far, lowest first. You can think
of this as the equivalent of “goal difference” when ranking teams on equal points in the real-world.

### Data Source

The real-world league table data used to calculate each score is retrieved from [football-data.org](https://www.football-data.org/).

This data source is polled every 15 minutes as a cron task. At this point, the most recently made [Prediction](docs/domain-knowledge.md#entryprediction)
for each [Entry](docs/domain-knowledge.md#entry) is used to calculate the [Scored Prediction](docs/domain-knowledge.md#scoredentryprediction)
for the current Game Week. It is also used to determine if a new Game Week has begun, or if the final Game Week has
ended (at which point the [Season](docs/domain-knowledge.md#season) is considered to be complete and is therefore finalised).

### FAQ

The Frontend also provides an FAQ page which can be customised on a per-[Realm](docs/domain-knowledge.md#realm) basis. 


## Domain Knowledge

For more details around core business logic and entities, see [Domain Knowledge](docs/domain-knowledge.md)


## New Features and Improvements

For more details around the project's intended Roadmap, see [New Features and Improvements](docs/new-features-and-improvements.md)


## Contributing

If you'd like to help out with this project in any way, please feel free to fork it and submit your PRs! 😁
