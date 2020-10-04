# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2020-10-04
### Added
- Transactional emails representing the opening and closing of a Season's Prediction Window.
- Cron jobs that check for recently opened, or forthcoming closing, Prediciton Windows within a 24-hour period, and issue
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
