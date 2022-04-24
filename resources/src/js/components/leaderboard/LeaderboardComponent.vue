<template>
    <div class="leaderboard-container">
        <div id="scoredEntryPredictionModal" class="modal fade" tabindex="-1" role="dialog">
            <div class="modal-dialog" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">{{focusedEntryNickname}} / Score: {{focusedEntryLeaderboardScore}}</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">
                        <round-navigation
                            :is-working="leaderboardsWorking"
                            :max-round-number="maxRoundNumber"
                            :round-number="baseRoundNumber"
                            v-on:decrement-round="prevRound"
                            v-on:increment-round="nextRound"
                        ></round-navigation>
                        <scored-entry-prediction
                            v-bind:entry-id="focusedEntryId"
                            v-bind:round-number="focusedRoundNumber"
                            v-bind:teams="teams"
                        ></scored-entry-prediction>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
                    </div>
                </div>
            </div>
        </div>
        <round-navigation
            :is-working="leaderboardsWorking"
            :max-round-number="maxRoundNumber"
            :round-number="baseRoundNumber"
            v-on:decrement-round="prevRound"
            v-on:increment-round="nextRound"
        ></round-navigation>
        <transition name="fade">
            <div v-if="leaderboardsErrorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                <button type="button" class="close" v-on:click="resetLeaderboardsErrorMessages">&times;</button>
                <p v-for="msg in leaderboardsErrorMessages" v-html="msg"></p>
            </div>
        </transition>
        <div v-if="focusedLeaderboard.rankings.length > 0" class="leaderboard-render-wrapper">
            <div v-if="leaderboardsWorking" class="loader-container">
                <img alt="loader" src="/assets/img/loader-light-bg.svg" />
            </div>
            <div v-if="!leaderboardsWorking && focusedLeaderboard.lastUpdated" class="last-updated text-center">Last updated on {{lastUpdatedVerbose}}</div>
            <table class="leaderboard-render rankings clickable">
                <thead>
                <tr>
                    <td colspan="4"></td>
                    <td class="text-right text-highlight">
                        Pts
                    </td>
                    <td class="text-right text-lolight">
                        Rnd
                    </td>
                    <td class="text-right text-lolight">
                        Min
                    </td>
                </tr>
                </thead>
                <tbody>
                <tr v-for="ranking in focusedLeaderboard.rankings" v-on:click="showScoredEntryPrediction(ranking.id, ranking.score)" class="leaderboard-row rankings-row">
                    <td class="position">
                        {{ranking.position}}
                    </td>
                    <td class="movement" v-html="getMovementMarkup(ranking.movement)"></td>
                    <td class="popout text-highlight">
                        <i class="fa fa-external-link" aria-hidden="true"></i>
                    </td>
                    <td class="name text-highlight">
                        {{entries[ranking.id]}}
                    </td>
                    <td class="text-right text-highlight">
                        {{ranking.total_score}}
                    </td>
                    <td class="text-right text-lolight">
                        {{ranking.score}}
                    </td>
                    <td class="text-right text-lolight">
                        {{ranking.min_score}}
                    </td>
                </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<script>
    const axios = require('axios').default
    const leaderboardPreloadBuffer = 2

    export default {
        name: 'LeaderBoard',
        props: {
            initialRoundNumber: {
                type: Number
            },
            rawEntries: {
                type: String
            },
            rawRankings: {
                type: String
            },
            rawTeams: {
                type: String
            },
            seasonId: {
                type: String
            },
            unix: {
                type: String
            }
        },
        data: function() {
            let entries = this.rawEntries === "" ? [] : JSON.parse(this.rawEntries)
            let teams = this.rawTeams === "" ? [] : JSON.parse(this.rawTeams)

            return {
                baseRoundNumber: this.initialRoundNumber, // round number that is inc/decremented by navigation controls
                entries, // map of entry id to entry nickname
                focusedEntryId: "", // entry id to display scored entry for
                focusedRoundNumber: this.initialRoundNumber, // round number to display leaderboard (and focused entry's scored entry) for
                leaderboards: this.populateInitialLeaderboard(), // map of leaderboards indexed by round number
                leaderboardsErrorMessages: [], // error messages relating to retrieval of leaderboard
                leaderboardsWorking: false, // denotes whether leaderboards are in the process of being retrieved
                maxRoundNumber: this.initialRoundNumber, // maximum available round number
                teams, // array of team objects with the schema id, name, short_name, crest_url
            }
        },
        methods: {
            getFocusedLeaderboard: function() {
                return this.leaderboards[this.focusedRoundNumber]
            },
            getMovementMarkup: function(movement) {
                if (movement > 0) {
                    return '<span class="movement-up"><i class="fas fa-caret-up"/></span>'
                }
                if (movement < 0) {
                    return '<span class="movement-down"><i class="fas fa-caret-down"/></span>'
                }
                return '<span class="movement-none"><i class="fas fa-minus"></i></span>'
            },
            nextRound: function() {
                this.baseRoundNumber++
            },
            populateInitialLeaderboard: function() {
                let parsedRankings = this.rawRankings === "" ? [] : JSON.parse(this.rawRankings)
                let lastUpdatedUnix = null
                if (this.unix !== "0") {
                    lastUpdatedUnix = this.unix + '000'
                }

                return this.populateLeaderboardByRound(
                    {},
                    this.initialRoundNumber,
                    parsedRankings,
                    lastUpdatedUnix
                )
            },
            populateLeaderboardByRound: function(leaderboards, roundNumber, rankings, lastUpdatedUnix) {
                let lastUpdated = lastUpdatedUnix != null ? new Date(parseFloat(lastUpdatedUnix)) : null
                leaderboards[roundNumber] = {rankings, lastUpdated}
                return leaderboards
            },
            prevRound: function() {
                this.baseRoundNumber--
            },
            resetLeaderboardsErrorMessages: function() {
                this.leaderboardsErrorMessages = []
            },
            retrieveLeaderboard: function(roundNumber, setFocus) {
                let component = this
                if (typeof component.leaderboards[roundNumber] != "undefined") {
                    if (setFocus) {
                        component.focusedRoundNumber = roundNumber
                    }
                    return
                }
                if (setFocus) {
                    component.leaderboardsWorking = true
                }
                axios.request({
                    method: 'get',
                    url: `/api/season/${this.seasonId}/leaderboard/${roundNumber}`
                }).then(function(response) {
                    let data = response.data.data
                    let rankings = data.leaderboard.rankings
                    let lastUpdatedUnix = Date.parse(data.leaderboard.last_updated)
                    component.setLeaderboard(roundNumber, rankings, lastUpdatedUnix)
                    if (setFocus === true) {
                        component.focusedRoundNumber = roundNumber
                    }
                }).catch(function(error) {
                    console.log(error)
                    if (setFocus) {
                        component.leaderboardsErrorMessages = ['Something went wrong :(<br />Please try again later']
                    }
                }).finally(function() {
                    component.leaderboardsWorking = false
                })
            },
            retrieveLeaderboards: function(lower, upper, focus) {
                for (let roundNumber = lower; roundNumber <= upper; roundNumber++) {
                    if (roundNumber < 1 || roundNumber > this.maxRoundNumber) {
                        continue
                    }
                    let setFocus = (roundNumber === focus)
                    this.retrieveLeaderboard(roundNumber, setFocus)
                }
            },
            setLeaderboard: function(roundNumber, rankings, lastUpdatedUnix) {
                this.leaderboards = this.populateLeaderboardByRound(
                    this.leaderboards,
                    roundNumber,
                    rankings,
                    lastUpdatedUnix
                )
            },
            showScoredEntryPrediction: function(entryID, score) {
                this.focusedEntryId = entryID
                $('#scoredEntryPredictionModal').modal('show')
            }
        },
        computed: {
            focusedEntryNickname: function() {
                return this.entries[this.focusedEntryId]
            },
            focusedEntryLeaderboardScore: function() {
                let rankings = this.getFocusedLeaderboard().rankings
                for (let idx in rankings) {
                    let ranking = rankings[idx]
                    if (ranking.id === this.focusedEntryId) {
                        return ranking.score
                    }
                }
                return 0
            },
            focusedLeaderboard: function() {
                return this.getFocusedLeaderboard()
            },
            lastUpdatedVerbose: function() {
                const helpers = require('../../helpers.js')
                if (this.focusedLeaderboard.lastUpdated === null) {
                    return ""
                }
                return helpers.formatVerboseDate(this.focusedLeaderboard.lastUpdated)
            },
        },
        watch: {
            baseRoundNumber: function(newRoundNumber) {
                this.resetLeaderboardsErrorMessages()
                let lower = newRoundNumber - leaderboardPreloadBuffer
                let upper = newRoundNumber + leaderboardPreloadBuffer
                this.retrieveLeaderboards(lower, upper, newRoundNumber)
                // TODO - retrieve scored entry predictions based on new round number
            },
        },
        mounted() {
            // preload initial number of "buffer" leaderboards
            let lower = this.maxRoundNumber-leaderboardPreloadBuffer
            this.retrieveLeaderboards(lower, this.maxRoundNumber)
        }
    }
</script>
