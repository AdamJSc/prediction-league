# New Features

[Back to README](../README.md)

Here is a brief outline of some new features that could form part of the project's Roadmap.

* Implement additional channels for transactional events, such as SMS or rich-text/HTML emails.


# Improvements

Here is a brief outline of some improvements that could be made to the system's existing implementation.

The majority of these have arisen due to time-constraints prior to the initial project roll-out, which had to be
launched in time for the 2020/21 football season!

## Testing with Season and Team Collections

* Work has been done to pass `SeasonCollection` and `TeamCollection` as dependencies to any domain agents/workers as required.
This replaces the previous implementation of a global data store for each one.

* However, for convenience the testsuites were updated to simply call `domain.GetSeasonCollection()` and assign the
first returned value as a `testSeason`. A similar approach has been taken with `domain.GetTeamCollection()`, so this data
is still effectively leveraging a global data store, instead of setting Season and Team expectations per test case.

* It would be useful to separate the test case logic properly, such that Season and Team expectations are explicitly defined
at the top of each test case for clarity.

## Email Retries

* The issuing of informational emails, such as "round complete", is attempted via Mailgun only once.

* Should this attempt fail, the send is currently not reattempted - although, a user's experience is unlikely to be
significantly impaired by lacking this information (it's more of a "nice to know").

* However, the same is also true of transactional emails, such as "new entry" confirmations. The non-receipt of this 
type of email **will** prevent the user from engaging with the application.

* A retry mechanism should be built such that any emails whose send attempt fails are retried a maximum of 3 times in
total with an increasing cool-off period of several seconds occurring between each attempt.

* Any emails that fail on all 3 attempts should be sent to a "dead letter" queue, which persists external to the
existing in-memory queue/channel so that appropriate subsequent action can be taken.

## Tokens

* For tokens that are left to expire without being "consumed" (i.e. an authenticated session), consider implementing a
cron job that cleans up all tokens whose `expires_at` timestamp has already elapsed.

* (✅ implemented as `RedeemedAt` in v2.1.0) For tokens that are "consumed" (i.e. a Short Code reset magic link), consider implementing a `consumed_at` timestamp
rather than removing them altogether.

## Payment

* Currently takes place exclusively via the Frontend using PayPal's [Basic Checkout Integration](https://developer.paypal.com/docs/checkout/integrate/).

* This means that [Entries](domain-knowledge.md#entry) must be approved manually by an Admin, in order that the
supplied _Payment Reference_ value can be verified as referring to a legitimate payment.

* Consider replacing the payment workflow such that payments can be handled exclusively on the Backend. This will negate
 the need for a manual approval step. 

## Registration Workflow

* Consider using a session/cookie-based workflow that will enable a user who is mid-signup to return to their most
recent stage of the sign-up workflow should they encounter any issues.

* This will enable a user to return to the payment processing stage of the sign-up workflow, should an issue occur
after the first [Entry](domain-knowledge.md#entry) creation stage has been successful.

## Cookies / Logout

* When user "logs out" after making a prediction, consider explicitly deleting the token on the backend rather than
just removing from frontend storage.

## Concurrency within Scheduled Tasks

* In the cron job that retrieves latest standings, consider generating scores concurrently (i.e. where
`domain.GenerateScoredEntryPrediction()` is invoked within a loop).

* This process only runs once every 15 minutes anyway and scale is expected to be very minimal, but performance benefits
are always fun regardless.

## Guard

* [Guards](domain-knowledge.md#guard) are an arbitrary permissions-based mechanism that provide agent methods with
the optional ability to determine whether or not an operation should be permitted.

* Guards belong to a custom context and must have their "attempt" value set within the request handler, such that the
agent method has the option of checking whether this is a match for a known "target" value.

* Perhaps consider something more robust/generic - e.g. an interface with a `Permit()` method that accepts a request object
and returns a `bool`. An implementation of this interface can then be passed to an agent method that requires the check.

## Vue integration

* Implementing Vue as a Node dependency causes Webpack to generate errors when building its assets, citing that the 
resultant source file "chunks" are too large.

* As a quick workaround for this, Vue is instead loaded from a CDN via a Frontend script tag. However, this means that
Vue will **only** run in one of development or production mode.

* Switching between the two is no longer a case of changing Webpack's build flag, and instead requires changing the script
tag within the HTML template itself.

* Consider resolving this by either...

    * ...implementing a boolean `.env` variable called `IS_PRODUCTION` which switches the Vue CDN URL that is loaded
    in the Frontend...
    * ...or resolving the Webpack chunking issue.

## Separate Unit/Integration Testsuites

* Whilst great care has been taken to ensure that the testsuite is as thorough as possible, the majority of tests are
essentially integration tests.

* This is due to the fact that many agent methods are effectively conduits for basic CRUD operations, so mocking the
repositories here would remove any meaningful business logic and therefore reduce the value of the test itself.

* Given that CI/CD for this project uses Bitbucket Pipelines, it is fortunately trivial to run integration tests as part
of the deployment pipeline since a MySQL container can be configured as a pipeline dependency.

* However, for the sake of sound engineering practice, consider writing separate unit and integration testsuites which
can be run independently of each other.
