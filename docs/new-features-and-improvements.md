[Back to README](../README.md)

# New Features

Here is a brief outline of some new features that could form part of the project's Roadmap.

* Implement additional means of communication for transactional events, such as SMS or rich-text/HTML emails.

* Implement transactional emails for additional events, such as notifying users when a new
[Prediction](domain-knowledge.md#entryprediction) window opens, or is about to close.


# Improvements

Here is a brief outline of some improvements that could be made to the system's existing implementation.

The majority of these have arisen due to time-constraints prior to the initial project roll-out, which had to be
launched in time for the 2020/21 football season!

## Global Data Store

* The [Seasons](domain-knowledge.md#season) and [Teams](domain-knowledge.md#team) data available to the system is 
configured in code as a central in-memory data store (see `datastore.Seasons` and `datastore.Teams` respectively).

* This is a deliberate design decision, in order that this data (which changes only once a year, before a real-world
season begins) can be carefully managed by the project maintainer, and does not require CRUD exposure from an end-user
perspective.

* However, this also means that agent methods must access these global data stores directly. This unfortunately prevents
the ability to mock [Season](domain-knowledge.md#season) and [Team](domain-knowledge.md#team) data when running tests
(particularly unit tests).

* The tests themselves must therefore access this global data directly when performing any setup within the test bootstrap.

* A much cleaner solution would instead require that the app's dependency container comprises a
"[Season](domain-knowledge.md#season) Data Provider" and a "[Team](domain-knowledge.md#team) Data Provider".

* This would enable the agent methods to retrieve this data from the container instead, which means the providers can be
easily mocked on the test dependency container that is injected into the agent constructors.

## Email Retries

* The issuing of informational emails, such as "round complete", is attempted via Mailgun only once.

* Should this attempt fail, the send is currently not reattempted - although, a user's experience is unlikely to be
significantly impaired by lacking this information (it's more of a "nice to know").

* However, the same is also true of transactional emails, such as "new entry" confirmations and Short Code resets. The
non-receipt of this type of email **will** prevent the user from engaging with the application.

* A retry mechanism should be built such that any emails whose send attempt fails are retried a maximum of 3 times in
total with an increasing cool-off period of several seconds occurring between each attempt.

* Any emails that fail on all 3 attempts should be sent to a "dead letter" queue, which persists external to the
existing in-memory queue/channel so that appropriate action can be taken at a later date.

## Short Codes

* A Short Code is a 6-character alphanumeric string that is affiliated with an [Entry](domain-knowledge.md#entry).

* It is randomly-generated at the point an [Entry](domain-knowledge.md#entry) is created, and confirmed to the end-user
within the "new entry" transactional email.

* The initial intention behind the Short Code was to serve as a quick means of authentication for operations that will
only occur a handful of times throughout a season (i.e. a user changing their [Prediction](domain-knowledge.md#entryprediction)).

* The likelihood of users' data being compromised is therefore reduced in so much as there is no risk that a user will
re-use one of their existing passwords from another (more data-sensitive) application, if we were to require them to
set this themselves.

* Also, the Short Code cannot be used to access or modify personal details such as name or email address either, so the decision
was made to store this in plain-text for convenience.

* However, as development of the system has progressed, the lines between a Short Code and a traditional password (within
the context of this project) have become somewhat ambiguous.

* Given the infrequency by which a user will actually need to make use of their Short Code, it is unlikely that they will
remember what their Short Code actually is when they come to need it. If they have deleted, or are unable to locate,
their "new entry" confirmation email, they have no way of retrieving their existing Short Code.

* For this reason, a Short Code Reset workflow has been developed, whereby a user can generate themselves a new random
Short Code via a series of double opt-in "magic link" emails - a bit like the way in which a traditional password may
be reset...

* This brings about a couple of issues in particular:

    * The new Short Code remains stored against its [Entry](domain-knowledge.md#entry) beyond its initial
    "consumption"/first usage - until such time as it is reset again by the user. In effect, this is acting like a
    traditional password - except for the fact that it is still present in the user's reset confirmation email, which is
    in no way obfuscated...

    * Resetting a Short Code overrides the previous one, which is discarded entirely. This means that at some point in
    the future, any user could **theoretically** be randomly-generated the exact same Short Code that was previously used
    by a different user (who has perhaps even entered the game via a different [Realm](domain-knowledge.md#realm) and
    [Season](domain-knowledge.md#season)).

* Therefore, a decision should be made as to whether Short Codes are retained longer-term, or replaced by a traditional
password mechanism.

    * If Short Codes are to be retained, they should become **ephemeral** - i.e. single-use for a limited timeframe, and
    discarded as soon as they have been consumed.
    
    * Alternatively, traditional passwords will require the usual hashing, additional considerations around storage and
    security of the system's data source, and sensible UX workflows that enable a user to set/reset their own password
    at any time (and not just during the pre-determined timeframe windows that allow a [Prediction](domain-knowledge.md#entryprediction)
    to be updated).

## Tokens

* For tokens that are left to expire without being "consumed" (i.e. an authenticated session), consider implementing a
cron job that cleans up all tokens whose `expires_at` timestamp has already elapsed.

* For tokens that are "consumed" (i.e. a Short Code reset magic link), consider implementing a `consumed_at` timestamp
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

* When user "logs out" after making a prediction, consider clearing cookie on the backend (and deleting the token)
rather than the frontend

## Concurrency within Scheduled Tasks

* In the cron job that retrieves latest standings, consider generating scores concurrently (i.e. where
`domain.ScoreEntryPredictionBasedOnStandings()` is invoked within a loop).

* This process only runs once every 15 minutes anyway and scale is expected to be very minimal, but performance benefits
are always fun regardless.

## Guard

* [Guards](domain-knowledge.md#guard) are arbitrary permissions-based mechanisms that provide agent methods with
the optional ability to determine whether or not an operation should be permitted.

* Guards belong to a custom context and must have their "attempt" value set within the request handler, such that the
agent method has the option of checking whether this is a match for a known "target" value.

* Perhaps consider something more robust/generic - e.g. an interface with a `Permit()` method that accepts a request object
and returns a `bool`. An implementation of this interface can then be passed to an agent methods that requires the check.

## Vue integration

* Implementing Vue as a Node dependency causes Webpack to generate errors when building its assets, citing that the 
resultant source file "chunks" are too large.

* As a quick workaround for this, Vue is instead loaded from a CDN via a Frontend script tag. However, this means that
Vue will run in **either** development or production mode.

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

* However, for the sake of sound engineering practice, consider enabling unit and integration testsuites to be run
independently of each other.
