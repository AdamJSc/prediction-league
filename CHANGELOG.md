# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

// TODO: feat - update CHANGELOG

### Added

- Demo data seeder cmd entrypoint
- Introduced new entities `MatchWeekSubmission`, `MatchWeekStandings` and `MatchWeekResult`.
- Config for 2022/23 Premier League season

### Changed

- Amended scoring mechanism:
    - Players still acquire 1 point for every position they have placed each team incorrectly in the rankings.
    - However, instead of this tally representing the player's full Match Week score, each player now starts the
      Match Week with 100 points and their calculated tally is **deducted** from this starting value
      (essentially taking a points "hit").
    - The aim of the game is now to accrue the **MOST** points across a season, not the fewest.
- Inverted leaderboard logic to reflect new aim of the game outlined above.
- General tidy up
    - Reconfigure Realm object
    - Amend site copy
    - Remove unused constants/symbols

### Fixed

- Team logos/crests render with consistent height

### Removed

- Requirement to enter Realm PIN on signup
- Legacy data seeder component

## [2.2.1] - 2022-06-15

### Changed

- Update colour scheme and branding
- Amend config to take running version and timestamp from build flags instead of env vars
- Update CI build

## [2.2.0] - 2022-04-26

### Added
- Render visual indication of leaderboard movement (the number of positions that each entry has moved up or down
based on the previous game week's rankings).
- Pop-out from Leaderboard which shows an individual user's scored prediction for a single game week now enables navigation 
through user's previous game week results within the same view. 

### Changed
- Preload/buffer retrieval of leaderboards to improve load times for user.
- Aesthetic improvements to leaderboard and scored entry views

### Fixed
- Extend duration of Edit Prediction token from 10 to 60 minutes, to prevent common timeout issues
- Typo in email address of seeded user

### Removed
- Verbosity flag on test run commands

## [2.1.5] - 2021-08-18

### Fixed
- Bug when adding a Prediction to an Entry. Check Season's timeframe for accepting Predictions instead of Entries.

## [2.1.4] - 2021-08-10

### Added
- Open Graph meta tags

## [2.1.3] - 2021-08-04

### Changed
- Minor copy changes

## [2.1.2] - 2021-07-29

### Added
- Release links to Changelog

### Changed
- Release script to include restarting of existing containers
- Home page copy

### Fixed
- FAQ accordion allows multiple expanded panels
- GitHub icon in page footer

## [2.1.1] - 2021-07-28

### Added
- Config for 2021/22 Premier League season

### Changed
- Minor web copy and design tweaks
- Tweaks to copy for communications

## [2.1.0] - 2021-07-27

### Added
- Additional token use-cases
- Ability to redeem a token
- Ability to purge non-redeemed tokens by user/type
- Ability for admin to generate an extended magic login link

### Changed
- Replace sequence of Season prediction windows, with a single window/timeframe.
- Implement ranking limit: permit a maximum of two teams to be swapped per Round/Game Week.
- Replace email/short code login with magic link

### Removed
- Deprecate Entry Short Code

## [2.0.0] - 2021-06-09

### Changed
- Reorganise significant elements of project architecture.
- Consolidate fragmented packages into `domain` and `adapters` layers.
- Compose service structs using explicit dependencies rather than single all-encapsulating injector.
- Re-implement cron and seeder as a runnable service component.
- Re-implement cron jobs against `worker` interface.
- Reduce required third-party Go modules.
- Improve error-wrapping for context.
- Improve logging.
    - Pass domain logger as dependency.
    - Implement logging levels.
- Tidy up tightly-coupled business logic.
- Implement clock interface dependency and replace "debug timestamp" with "frozen clock" instance.
- Improve test coverage.
- Update build and deployment pipeline.
- Update docs.

## [1.1.4] - 2021-05-23

### Fixed
- Accommodate deprecation of `standingsType` query param when retrieving Standings from upstream Football Data API.
Perform this filter explicitly, by iterating over response payload and checking `type` value of all Standings objects.  

## [1.1.3] - 2021-03-27

### Added
- Database startup error detail

### Fixed
- Docker image build copies additional source files not included in binary

## [1.1.2] - 2021-03-27

### Changed
- Docker builder pattern to reduce the Docker image size

### Security
- Bump Go from 1.14 to 1.16
- Bump ini from 1.3.5 to 1.3.8
- Bump axios from 0.19.2 to 0.21.1
- Bump elliptic from 6.5.3 to 6.5.4

## [1.1.1] - 2020-10-20

### Fixed
- Bug where the most recently created Scored Entry Prediction was not necessarily being retrieved for each specified
combination of Entry/Round Number (Game Week) when building the Leaderboard, due to spurious behaviour of MySQL's
Group/Order By functions.
- Amended the sub-queries that are used to select Entries' cumulative scores by Realm, such that the order of records
produced when descending by Scored Entry Prediction's `created_at` field is guaranteed to be retained by the parent query. 

## [1.1.0] - 2020-10-04

### Added
- Transactional emails representing the opening and closing of a Season's Prediction Window.
- Cron jobs that check for recently opened, or forthcoming closing, Prediction Windows within a 24-hour period, and issue
email comms to all Entrants of the respective active Season.

## [1.0.3] - 2020-09-27

### Security
- Bump elliptic from 6.5.2 to 6.5.3
- Bump lodash from 4.17.15 to 4.17.20
- Bump copy-webpack-plugin from 6.0.2 to 6.1.1
- Bump terser-webpack-plugin from 1.4.3 to 1.4.5
- Bump node-sass from 4.14.0 to 4.14.1
- Bump node-sass-chokidar from 1.4.0 to 1.5.0

## [1.0.2] - 2020-09-27

### Fixed
- When retrieving latest standings and checking for a completed season, fixed check so that this is made on the
Standings object that has just been received from data client.
- Add nil check when retrieving Realm from context.

## [1.0.1] - 2020-08-31

### Changed
- Duration of session token for updating a Prediction from 20 minutes to 60 minutes.

## [1.0.0] - 2020-08-29

### Added
- This project to the Open Source "dimension"...

[Unreleased]: https://github.com/AdamJSc/prediction-league/compare/v2.2.1...HEAD
[2.2.1]: https://github.com/AdamJSc/prediction-league/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/AdamJSc/prediction-league/compare/v2.1.5...v2.2.0
[2.1.5]: https://github.com/AdamJSc/prediction-league/compare/v2.1.4...v2.1.5
[2.1.4]: https://github.com/AdamJSc/prediction-league/compare/v2.1.3...v2.1.4
[2.1.3]: https://github.com/AdamJSc/prediction-league/compare/v2.1.2...v2.1.3
[2.1.2]: https://github.com/AdamJSc/prediction-league/compare/v2.1.1...v2.1.2
[2.1.1]: https://github.com/AdamJSc/prediction-league/compare/v2.1.0...v2.1.1
[2.1.0]: https://github.com/AdamJSc/prediction-league/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/AdamJSc/prediction-league/compare/v1.1.4...v2.0.0
[1.1.4]: https://github.com/AdamJSc/prediction-league/compare/v1.1.3...v1.1.4
[1.1.3]: https://github.com/AdamJSc/prediction-league/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/AdamJSc/prediction-league/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/AdamJSc/prediction-league/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/AdamJSc/prediction-league/compare/v1.0.3...v1.1.0
[1.0.3]: https://github.com/AdamJSc/prediction-league/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/AdamJSc/prediction-league/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/AdamJSc/prediction-league/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/AdamJSc/prediction-league/releases/tag/v1.0.0
