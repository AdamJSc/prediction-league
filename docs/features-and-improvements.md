[Back to README](../README.md)

TODO - this...

# Features

* Additional comms methods - HTML email + SMS
* Additional comms events - prediction window opens, prediction window pre-close reminder


# Improvements

## Guard

## Payment

* Verify via Backend

## Short Codes / Passwords

* Lines blurred as development has gone on
* Also, expired short codes aren't retained, so could **theoretically** be generated again in the future

## Tokens

* Robustness
* Clean up of expired tokens (`elapsed_at`)
* Set a `consumed_at` date instead of deleting

## Cookie Management

## Registration Workflow

* Return to payment stage mid-entry
* Use cookies

## Scheduled Tasks Concurrency

* Make processing of Standings concurrent when retrieving Latest Standings + scoring Entry Predictions
* Process runs once every 15 minutes and scale currently very small

## Queue retries

* If sending of email fails, retry it X times with a cool-off in between each

## Global Data Store

* Hard-coded seasons and teams data
* Replace with a "seasons data provider" and "teams data provider" that can be injected via container

## Vue integration

* Pulling from CDN prevents debug mode in development
* Webpack was warning against huge chunks otherwise

## Split out tests
