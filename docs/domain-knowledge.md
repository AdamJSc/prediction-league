# Domain Knowledge

[Back to README](../README.md)

## Entities

Here is a detailed breakdown of each Entity:

### Realm

* A `Realm` is an arbitrary "flag" that represents a distinct instance of the game. It can also be thought of as a sub-grouping
of [Entries](#entry).

* [Entries](#entry) that have different Realm values belong to different game instances, and are therefore not
competing against each other.

* The Realm Name is determined by the domain portion of the request URL via which an [Entry](#entry) is initially created.
This excludes any port numbers that are appended to the domain name portion of the request URL.

* This approach enables multiple hosts/domains/sub-domains to route traffic to a single running binary, which itself can
serve multiple game instances at the same time.

* Each Realm must comprise one [Season](#season), which determines the [Teams](#team) that are available to select when
making a [Prediction](#entryprediction), as well as the real-world league table data that each [Entry](#entry) will be
scored against.

* The [Season](#season) that is associated with a Realm can (and should) change every year, in line with changes to the
currently active real-world season - i.e. `201920_1` ("Premier League 2019/20") becomes `202021_1` ("Premier League 2020/21")
etc.

* Each Realm must also have a corresponding PIN that is supplied when creating an [Entry](#entry), in order to prevent
unwelcome sign-ups.

* Realms must be configured in advance, within the `./data/realms.yml` file. This file's payload comprises the single
root-level key named `realms`, and each key beneath this should match the lowercase realm name (domain) -
e.g. `localhost` or `my.sub.domain.com`. See the file itself for an example.

* Each Realm config object should comprise the following values:
    * `origin`: Base URL to use for links inside transactional emails that are sent to [Entries](#entry) belonging to Realm
    (e.g. `http://localhost:3000`)
    * `contact.name`: Name to sign off transactional emails with (e.g .`Harry R and the PL Team`)
    * `contact.email_proper`: Formatted support email address (e.g. `wont_you_please_please_help_me@localhost`)
    * `contact.email_sanitised`: Sanitised support email address for display purposes (e.g. `wont_you_please_please_help_me (at) localhost`)
    * `contact.email_do_not_reply`: Sender's email address on transactional emails (e.g. `do_not_reply@localhost`)
    * `sender_domain`: Outgoing sender domain validated via DNS with Mailgun.
    * `pin`: Arbitrary string value required when creating a new [Entry](#entry) to prevent unwanted sign-ups (e.g. `MYPIN1234`)
    * `season_id`: ID of current [Season](#season) that Realm pertains to, must be an existing key within the supplied `SeasonCollection` (e.g. `201920_1`)
    * `entry_fee.amount`: Numerical entry fee value charged via PayPal, currently supports GBP only (e.g. `1.23`)
    * `entry_fee.label`: Entry fee's display value (e.g. `£1.23`)
    * `entry_fee.breakdown`: Array of strings that explain breakdown of entry fee (e.g. `- £1.00 kitty contribution, - £0.23 processing fee`) 
    * `analytics_code`: Google Analytics code for Realm (optional, analytics ignored if left blank)

* The values of these payloads are parsed as `Realm` objects within the main app bootstrap, and subsequently retrievable
by accessing `realms.GetByName(realm_name)` on the app's container.

* The default Realm Name when running locally is `localhost`.

### Season

* A `Season` represents a real-world tournament (such as "Premier League 2020/21").

* The Seasons data used throughout the system is defined in code and instantiated within the app bootstrap (see `domain.GetSeasonCollection()`).

* This is deliberately controlled by the project maintainer as a one-off action that is required approximately once a year
(i.e. between one Season finishing and another beginning).

* For details on the system's default Season, see ["FakeSeason"](#fakeseason) (below).

### Team

* A `Team` represents a team that are competing within a [Season](#season).

* Each [Season](#season) defines a slice of Teams that...
    * ...are presented to the User when setting their [Prediction](#entryprediction), and...
    * ...must be present within the ingested real-world [Standings](#standings) from the upstream data source
    for a given [Season](#season).

### Entry

* An `Entry` represents a user who has been entered into a game instance.

* Each Entry belongs to a single combination of [Realm](#realm) and [Season](#season).

* Entries are only considered active if they have been "approved" by an Admin (see [Payments](../README.md#payments)).

* Each Entry comprises an auto-generated `ShortCode` which must be provided as a means of authentication
when "logging in" to make changes to a [Prediction](#entryprediction)

### EntryPrediction

* An `EntryPrediction` (also referred to as a `Prediction`) represents a Prediction that has been made as part of an [Entry](#entry)

* It comprises a creation timestamp, as well as a JSON array containing a sequence of [Team](#team) IDs that are considered
to be the chosen order of [Teams](#team) by which the [Entry](#entry) will accrue points for all current and subsequent
game weeks (until a new EntryPrediction is made).

* Each EntryPrediction belongs to a single [Entry](#entry).

* An EntryPrediction can only be created and not updated.

* The act of a user changing their existing EntryPrediction creates a new one with a more recent timestamp instead. This
facilitates an "event-based" paradigm, providing a full historic audit trail of all EntryPredictions that have ever
belonged to an [Entry](#entry).

### Standings

* A `Standings` object represents an instance of a real-world league table consumed from the upstream data source.

* It comprises a JSON object called Rankings, which represents a sequence of [Team](#team) IDs along with their meta values,
such as number of matches played, won, drawn, lost etc. for each [Team](#team).

* It is inflated as part of the [Retrieve Latest Standings](#retrieve-latest-standings) cron job.

* Each Standings record is unique by SeasonID and Round Number (game week) which is overwritten by each cron job execution,
until the next game week is reached (or the [Season](#season) reaches completion). 

### ScoredEntryPrediction

* A `ScoredEntryPrediction` object represents a [Prediction](#entryprediction) that has been provided with a score of
penalty points based on a given [Standings](#standings) object.

* It is unique to a single combination of a [Prediction](#entryprediction) object and a [Standings](#standings) object.

* It is used to calculate the total cumulative score for each [Entry](#entry) on a [LeaderBoard](#leaderboard).

### Ranking

* A `Ranking` represents an arbitrary ID that has an associated numerical position (i.e. an ordered position within a sequence).

### RankingWithMeta

* A `RankingWithMeta` represents a [Ranking](#ranking) that has an associated map of meta data, such as those found
for each [Team](#team) within a [Standings](#standings) object.

### RankingWithScore

* A `RankingWithScore` represents a [Ranking](#ranking) that has an associated numerical position, such as those found
within a [ScoredEntryPrediction](#scoredentryprediction) object.

### RankingWithScore + MetaPosition

* This object represents a [RankingWithScore](#rankingwithscore) object that has an associated "meta position", such as
those found within a `scoredEntryPredictionResponseRanking` object.

* Here, the "meta position" represents the real-world league table position of the Team ID represented by the Ranking, e.g.
    * For the following pseudo `Standings.Rankings`:
        * `1st Team A`
        * `2nd Team B`
        * `3rd Team C`
    * ...and the following pseudo `EntryPrediction.Rankings`:
        * `1st Team B`
        * `2nd Team C`
        * `3rd Team A`
    * ...the following pseudo-`RankingWithScore + MetaPosition` would result:
        * `1st Team B (Score = 1, Meta Position = 2)`
        * `2nd Team C (Score = 1, Meta Position = 3)`
        * `3rd Team A (Score = 2, Meta Position = 1)`

### LeaderBoard

* A `LeaderBoard` represents the cumulative total scores for all [Entries](#entry) within a [Season](#season), ordered
by total score (lowest first), then by minimum score (lowest first), then by current round score (lowest first).

* It is unique to a Round Number (game week).

### LeaderBoardRanking

* A `LeaderBoardRanking` represents the position of a single [Entry](#entry) within a [LeaderBoard](#leaderboard).

* It combines a [RankingWithScore](#rankingwithscore), with a numerical Total Score and a numerical Min Score.

### Token

* A `Token` comprises an ID (an arbitrary string of 32 alphanumeric characters) as well as a Value
(a string that represents some other existing entity).

* Each Token has an `IssuedAt` timestamp, representing the time at which it was issued.

* Each Token also has an `ExpiresAt` timestamp, representing the point at which it can no longer be considered valid.

* Tokens have one of two types:
    * `Auth` Tokens are used to identify a session cookie. They have a duration of 20 minutes before expiring and their
    Value represents the [Entry](#entry) ID associated with the session.
    * `ShortCode Reset` Tokens are used to denote an in-progress ShortCode reset "magic link" workflow. They have a
    duration of 10 minutes before expiring and their Value represents the [Entry](#entry) ID associated with ShortCode reset.

* At the moment, Tokens are in no way flagged as "consumed". There is no clean-up process for Tokens that have expired, and 
ShortCode Reset Tokens are simply removed after having been used. This should ideally be reviewed at some point in the future.

### Email

* An `Email` represents the content and meta data of an email message to be issued via Mailgun.

### MessageIdentity

* A `MessageIdentity` represents the Name and Address of an individual [Email](#email) message sender or recipient.


## Other Business Logic

### Retrieving Latest Standings

Real-world league table standings are consumed from the [football-data.org](https://www.football-data.org/) API, for the
purposes of calculating new scores for each [Entry](#entry) per game week (also known as
[ScoredEntryPredictions](#scoredentryprediction)).

This occurs within a cron job that is scheduled to run every 15 minutes (see `scheduler.newRetrieveLatestStandingsJob()`).

A separate cron job is instantiated for every existing [Season](#season) that fulfils all of the following criteria
at the time the service boots up:

* ...is currently affiliated with an existing [Realm](#realm), and...
* ...comprises a `ClientID` that is not `nil`.

The `.env` variable `FOOTBALLDATA_API_TOKEN` must also have a value for any of these cron jobs to run.

The cron job's task executes the following logic:

* Check that [Season](#season) is `Active` - exit if not.

* Retrieve **newest** [EntryPredictions](#entryprediction) associated with the provided [Season](#season) - exit if none.

* Retrieve latest [Standings](#standings) from football data client - exit if error

* Validate and sort retrieved [Standings](#standings) - exit if error

* Check for an existing [Standings](#standings) record that matches the Round Number (game week) of the retrieved
[Standings](#standings):
    * If we already have it, just update its Rankings
    * Otherwise we have a new one, so retrieve the [Standings](#standings) for the **previous** Round Number (game week) 
    mark it as "finalised", and continue with this one instead.

* For each [Prediction](#entryprediction) we retrieved earlier:
    * Calculate new [ScoredEntryPrediction](#scoredentryprediction) based on our [Standings](#standings)
    * Insert new/Update existing [ScoredEntryPrediction](#scoredentryprediction) for current [Entry](#entry) ID and
    [Standings](#standings) ID.

* If [Standings](#standings) has been marked as finalised, then issue a "round complete" email to each [Entry](#entry)

### "FakeSeason"

Much of the functionality within this system is time-sensitive in relation to the [Season](#season) that applies to the
current [Realm](#realm).

Functionality such as whether or not an [Entry](#entry) or [Prediction](#entryprediction)
can be created or updated at any given moment is determined by the corresponding timeframes set on the [Season](#season)
object itself.

Usually these will be **absolute** timeframes that pertain to the dates relevant to a real-world [Season](#season)
(see `201920_1` in the collection returned by `domain.GetSeasonCollection()` as an example).

For this reason, the default [Realm](#realm) (`localhost`) is affiliated with a [Season](#season) which has the ID `FakeSeason`
and whose sole purpose is to enable time-sensitive operations to be more easily debugged.

This [Season's](#season) timeframes are **relative** to the point at which the system is run, so that core functionality
can be used immediately when running locally.

The schedule takes place as follows:

* 0 mins - 20 mins
    * [Entries](#entry) **can** be created or updated
    * [Predictions](#entryprediction) **can** be created
* 20 mins - 40 mins
    * [Entries](#entry) **cannot** be created or updated
    * [Predictions](#entryprediction) **cannot** be created
* 40 mins - 60 mins
    * [Entries](#entry) **cannot** be created or updated
    * [Predictions](#entryprediction) **can** be created
* 60 mins+
    * [Entries](#entry) **cannot** be created or updated
    * [Predictions](#entryprediction) **cannot** be created

### Transactional Emails

When a new email is issued via one of the `domain.CommunicationsAgent` object methods, it is sent to a channel that
acts as an arbitrary message queue.

This channel is consumed within a separate long-running _goroutine_ which is responsible for the physical dispatch of the
email itself via Mailgun, in an attempt to leverage decoupling/concurrency.

Due to time constraints, there is currently no retry/cool-off mechanism or dead-letter handling for failed emails.
However, this should be revisited in the future.

### Adding a new Season

To add a new [Season](#season), define an additional `Season` struct within the map provided by the function `domain.GetSeasonsColection()`.

This struct must adhere to the validation rules found within `domain.ValidateSeason()`.

Running the testsuite again will apply the rules to each `Season` in the map and fail if any aren't met.

### Guard

A Guard (see `domain.Guard`) is an arbitrary means of assessing whether an agent method should continue its execution,
and can be thought of as a very basic and rudimentary per-method permission system.

There are two component parts to a Guard: the `Attempt` and the `Target`.

The `Attempt` is usually the value supplied within a request in some way.

The `Target` is usually a "known" value that the `Attempt` must equal in order that the agent method experiences a 
positive match.

For example...

Most route handlers will at some point invoke `ctx := contextFromRequest(r, c)`, where `r` is a HTTP Request object,
and `c` is the container comprising all of our app's dependencies.

This will return a new context with two values added: one is an empty `Guard` object, and the other is a `Realm` object
that has been populated with details of the request's current [Realm](#realm).

The route handler might then invoke `domain.GuardFromContext(ctx).SetAttempt(input.PIN)` to set the Guard's Attempt value -
i.e. the value "attempting" to match the Target - using some data point that has been supplied in the user's request

(In this case, our Guard Attempt is the "PIN" field of the incoming request body, so our agent method will likely want
this to match the PIN of the request's current [Realm](#realm) so it can decide to allow the operation to continue).

The handler will pass this context to the main agent method that it invokes (e.g. `domain.EntryAgent.CreateEntry(ctx ....)`).

The agent method can then invoke `domain.GuardFromContext(ctx).AttemptMatchesTarget(ctx.Realm.PIN)` which returns
a `boolean` to determine whether or not the incoming request should be authorised.
