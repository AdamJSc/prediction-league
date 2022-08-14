# Domain Knowledge

[Back to README](../README.md)

## Entities

Here is a detailed breakdown of each Entity:

### Realm

* A `Realm` represents a distinct instance of the game. It can also be thought of as a sub-grouping of [Entries](#entry).

* Entries that have different Realm values belong to different game instances, and are therefore not competing against
each other.

* The Realm Name is determined by the domain portion of the request URL via which an Entry is initially created.
This excludes any port numbers that are appended to the domain name portion of the request URL.

* This approach enables several hosts/domains/sub-domains to be proxied to the same hosted application, therefore 
accommodating multiple game instances at the same time.

* Each Realm must comprise one [Season](#season), which determines the [Teams](#team) that are available to players when
making a [Prediction](#entryprediction), as well as the real-world league table data that each Prediction will be
scored against.

* The Season that is associated with a Realm can (and should) change every year, in line with changes to the currently
active real-world season - i.e. `201920_1` ("Premier League 2019/20") becomes `202021_1` ("Premier League 2020/21")
etc.

* A Realm has historically had a corresponding PIN that must be supplied when creating an Entry, in order to prevent
unwelcome sign-ups (players must be given the PIN by whoever is administering the game instance).
However, this feature was "deprecated" in v2.3.0 (the PIN mechanism has been retained, but the frontend simply includes
its value as part of the Entry creation request).

* Realms must be configured in advance, within the `./data/realms` directory. Each Realm is provided as a sub-directory
which contains the files `main.yml` and `faqs.yml`. See the `localhost` example in this repo for the exact schema required
for these files.

* The values of these payloads are parsed as `Realm` objects within the main app bootstrap, and subsequently retrievable
by accessing `GetByName(realm_name)` on the `RealmCollection` which originates in the app's container and is passed as a dependency
to each domain entity that requires it, such as handlers, agents, workers etc.

* The default Realm Name when running locally is `localhost`, so please ensure that you are issuing API requests to the
 base URI `http://localhost` instead of any other alias such as `http://127.0.0.1` etc.

### Season

* A `Season` represents a real-world tournament (such as "Premier League 2020/21").

* The Seasons data used throughout the system is defined in code and instantiated on the app container as `SeasonCollection`.
Again, this is passed as a dependency to each domain entity that requires it, such as handlers, agents, workers etc.

* This data is deliberately controlled by the project maintainer as a one-off action since updating it is required
approximately once a year (i.e. between one Season finishing and another beginning) and applies to the project as a whole.

* For details on the system's default Season, see ["FakeSeason"](#fakeseason) (below).

### Team

* A `Team` represents a team that competes within a [Season](#season).

* Each Season defines a slice of Teams that...
    * ...are presented to the player when setting their [Prediction](#entryprediction), and...
    * ...must be present within the ingested real-world [Standings](#standings) from the upstream data source
    for a given [Season](#season).

* The Teams data used throughout the system is defined in code and instantiated on the app container as `TeamCollection`.
Again, this is passed as a dependency to each domain entity that requires it, such as handlers, agents, workers etc.

* This data is deliberately controlled by the project maintainer as a one-off action since updating it is required
  approximately once a year (i.e. between one Season finishing and another beginning) and applies to the project as a whole.

### Entry

* An `Entry` represents a player who has been entered into a game instance.

* Each Entry belongs to a single combination of [Realm](#realm) and [Season](#season).

* Entries are only considered active if they have been "approved" by an Admin (see [Payments](../README.md#payments)).
Prior to this, they are not included within the [Leaderboard](#leaderboard).

* To login and make a [Prediction](#entryprediction), the user must generate a magic login link (see [Token](#token)).

### EntryPrediction

* An `EntryPrediction` (also referred to as a `Prediction`) represents a Prediction that has been made as part of an [Entry](#entry).

* It comprises a creation timestamp, as well as a JSON array containing a sequence of [Team](#team) IDs that the player
will be scored against each [Match Week](../README.md#match-weeks).

* Each EntryPrediction belongs to a single [Entry](#entry).

* An EntryPrediction can only be created and not updated.

* The act of a user changing their existing EntryPrediction creates a new one with a more recent timestamp instead. This
facilitates an "event-based" paradigm, providing a full historic audit trail of all EntryPredictions that have ever
belonged to an Entry.

* Only one EntryPrediction per Entry is ever considered to be active/current at any given time.

* The most recently-created EntryPrediction is used to generate the [Scored Prediction](#scoredentryprediction) for the current
and future Match Weeks (until a new EntryPrediction is created).

### MatchWeekSubmission

* Introduced in v2.3.0 to accommodate new scoring rules.

* Transformed from `EntryPrediction` entity within `GenerateScoredEntryPrediction` method.

* Interim state is stored in database by this method before being discarded to accommodate potential future migrations.

### Standings

* A `Standings` object represents an instance of a real-world league table consumed from the upstream data source.

* It comprises a JSON object called Rankings, which represents a sequence of [Team](#team) IDs along with their meta values,
such as number of matches played, won, drawn, lost etc. for each Team.

* It is inflated as part of the [Retrieve Latest Standings](#retrieve-latest-standings) cron job.

* Each Standings record is unique by SeasonID and Round Number (Match Week) which is overwritten by each cron job execution,
until the next Match Week is reached (or the [Season](#season) reaches completion). 

### MatchWeekStandings

* Introduced in v2.3.0 to accommodate new scoring rules.

* Transformed from `Standings` entity within `GenerateScoredEntryPrediction` method.

### ScoredEntryPrediction

* A `ScoredEntryPrediction` object represents a [Prediction](#entryprediction) that has been provided with a score based
on a given [Standings](#standings) object.

* It is unique to a single combination of a Prediction and a Standings.

* It is used to calculate the total cumulative score for each [Entry](#entry) on a [Leaderboard](#leaderboard).

### MatchWeekResult

* Transformed from `ScoredEntryPrediction` within `GenerateScoredEntryPrediction` method.

* Interim state is stored in database by this method before being discarded to accommodate potential future migrations.

* Also introduces the concept of "Modifiers" which are the individual factors that influence the overall Score on a `MatchWeekResult`.

* Modifiers are stored with an arbitrary Code so that their nature (description etc.) can be recalled in the future,
as well as a Value (the amount by which to affect the Score so that the modifiers can be "replayed" as required).

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
those found within a `scoredEntryPredictionResponseRanking` object from the upstream football data API response.

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

### Leaderboard

* A `Leaderboard` represents the cumulative total scores for all [Entries](#entry) within a [Season](#season), ordered
by total score (highest first), then by maximum score (highest first), then by current Match Week score (highest first).

* It is unique to a Round Number (Match Week).

### LeaderboardRanking

* A `LeaderboardRanking` represents the position of a single [Entry](#entry) within a [LeaderBoard](#leaderboard).

* It combines a [RankingWithScore](#rankingwithscore), with a numerical Total Score and a numerical Min Score.

### Token

* A `Token` comprises an ID (an arbitrary string of 32 alphanumeric characters) as well as a Value
(a string that represents some other existing entity, usually an [Entry](#entry)).

* Each Token has an `IssuedAt` timestamp, representing the time at which it was issued.

* Each Token also has a `RedeemedAt` timestamp, representing the time at which the token was consumed and rendered used.
This is `NULL`/`nil` for any instances that have been created but not yet been redeemed.

* Each Token also has an `ExpiresAt` timestamp, representing the point at which it can no longer be considered active.

* Tokens represent one of four types:
    * `Auth` Tokens are used to identify a user's session. They have a duration of 60 minutes before expiring and their
    Value represents the [Entry](#entry) ID associated with the session.
    * `Entry Registration` Tokens are generated as single-use in order to facilitate the payment step that follows creating
    an Entry. They have a duration of 10 minutes before expiring and their Value represents the associated [Entry](#entry) ID.
    * `Magic Login` Tokens are used as part of a magic login link. They have a duration of 10 minutes before expiring
    and their Value represents the [Entry](#entry) ID associated with requested magic login.
    * `Prediction` Tokens are generated as single-use in order to facilitate the creation of a new Entry Prediction.
    They have a duration of 60 minutes before expiring and their Value represents the associated [Entry](#entry) ID.

* At the moment, there is no clean-up process for Tokens that have expired without being redeemed/consumed. A bulk-delete
 agent method has been implemented although is not yet invoked as part of a cron job etc. This should ideally be reviewed
 at some point in the future.

### Email

* An `Email` represents the content and meta data of an email message to be issued via Mailgun.

### MessageIdentity

* A `MessageIdentity` represents the Name and Address of an individual [Email](#email) message sender or recipient.


## Other Business Logic

### Retrieving Latest Standings

Real-world league table standings are consumed from the [football-data.org](https://www.football-data.org/) API, for the
purposes of calculating new scores for each [Entry](#entry) per Match Week (also known as
[ScoredEntryPrediction](#scoredentryprediction)).

This occurs as a cron job that is scheduled to run every 15 minutes (see `domain.RetrieveLatestStandingsWorker`).

A separate cron job is instantiated for every existing [Season](#season) that fulfils all of the following criteria
at the time the service boots up:

* ...is currently affiliated with an existing [Realm](#realm), and...
* ...comprises a `ClientID` (for calling the upstream football data API) that is not `nil`.

The `.env` variable `FOOTBALLDATA_API_TOKEN` must also have a value for any of these cron jobs to run.

The cron job's task executes the following logic:

* Check that the associated Season is `Active` - exit if not.

* Retrieve **newest** [Predictions](#entryprediction) belonging to the associated [Season](#season) - exit if none.

* Retrieve latest [Standings](#standings) from football data client - exit if error

* Validate and sort retrieved Standings - exit if error

* Check for an existing Standings object that matches the Round Number (Match Week) of the retrieved Standings:
    * If we already have it, just update its Rankings
    * Otherwise we have a new one, so retrieve the Standings for the **previous** Round Number (Match Week), mark it as
      "finalised", and continue with this one instead.

* For each Prediction we retrieved earlier:
    * Calculate new ScoredPrediction based on our Standings.
    * Upsert ScoredEntryPrediction for current Entry ID and Standings ID.

* If Standings has been marked as finalised, then issue a "round complete" email to each player (Entry).

### "FakeSeason"

Much of the functionality within this system is time-sensitive in relation to the [Season](#season) that applies to the
current [Realm](#realm).

Functionality such as whether or not an [Entry](#entry) or [Prediction](#entryprediction) can be created or updated at
any given moment is determined by the corresponding timeframes set on the Season object itself.

Usually these will be **absolute** timeframes pertaining to dates that are relevant to a real-world Season (see
`201920_1` in the collection returned by `domain.GetSeasonCollection()` as an example).

For this reason, the default Realm (`localhost`) is affiliated with a Season which has the ID `FakeSeason` and whose
sole purpose is to enable time-sensitive operations to be more easily debugged.

The `FakeSeason` timeframes are **relative** to the point at which the system is run, so that core functionality
can be used immediately when running locally.

The schedule takes place as follows:

* 0 mins - 20 mins
    * Entries **can** be created or updated
    * Predictions **can** be created
* 20 mins - 40 mins
    * Entries **cannot** be created or updated
    * Predictions **cannot** be created
* 40 mins - 60 mins
    * Entries **cannot** be created or updated
    * Predictions **can** be created
* 60 mins+
    * Entries **cannot** be created or updated
    * Predictions **cannot** be created

### Transactional Emails

When a new email is issued via one of the `domain.CommunicationsAgent` methods, it is sent to a channel that
acts as an arbitrary message queue.

This channel is consumed within a separate long-running _goroutine_ which is responsible for the physical dispatch of the
email itself via Mailgun, in an attempt to leverage decoupling/concurrency.

Due to time constraints, there is currently no retry/cool-off mechanism or dead-letter handling for failed emails.
However, this should be revisited in the future.

Also consider replacing the in-memory queue with a hosted instance of RabbitMQ/PubSub etc.

### Adding a new Season

To add a new [Season](#season), define an additional `Season` struct within the map provided by the function `domain.GetSeasonColection()`.

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
this to match the PIN of the request's current Realm so it can decide whether to allow the operation to continue).

The handler will pass this context to the main agent method that it invokes (e.g. `domain.EntryAgent.CreateEntry(ctx ....)`).

The agent method can then invoke `domain.GuardFromContext(ctx).AttemptMatchesTarget(ctx.Realm.PIN)` which returns
a `boolean` to determine whether or not the incoming request should be authorised.

This permissions-based mechanism can most likely be simplified and should be revisited in the future (replace with routing middleware etc.)
