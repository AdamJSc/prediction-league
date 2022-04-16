<template>
    <div class="leaderboard-container">
        <div id="scoredEntryPredictionModal" class="modal fade" tabindex="-1" role="dialog">
            <div class="modal-dialog" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">{{focusedEntry.nickname}} / Score: {{focusedEntry.score}} / Round: {{roundNumber}}</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">
                        <scored-entry-prediction v-bind:entry-id="focusedEntry.id" v-bind:round-number="roundNumber" v-bind:teams="teams"></scored-entry-prediction>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
                    </div>
                </div>
            </div>
        </div>
        <table class="round">
            <tr>
                <td class="round-navigation text-left"><i v-on:click="prevRound" v-if="!working && (roundNumber > 1)" class="round-navigation fa fa-chevron-left" aria-hidden="true"></i></td>
                <td class="text-center">Round {{roundNumber}}</td>
                <td class="round-navigation text-right"><i v-on:click="nextRound" v-if="!working && (roundNumber < maxRoundNumber)" class="fa fa-chevron-right" aria-hidden="true"></i></td>
            </tr>
        </table>
        <transition name="fade">
            <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                <p v-for="msg in errorMessages" v-html="msg"></p>
            </div>
        </transition>
        <div v-if="focusedLeaderboard.rankings.length > 0" class="leaderboard-render-wrapper">
            <div v-if="working" class="loader-container">
                <img alt="loader" src="/assets/img/loader-light-bg.svg" />
            </div>
            <div v-if="!working && focusedLeaderboard.lastUpdated" class="last-updated text-center">Last updated on {{lastUpdatedVerbose}}</div>
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
                <tr v-for="ranking in focusedLeaderboard.rankings" v-on:click="showScoredEntryPrediction(ranking.id, entries[ranking.id], ranking.score)" class="leaderboard-row rankings-row">
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
                entries,
                errorMessages: [],
                focusedEntry: {
                    id: "",
                    nickname: "",
                    score: 0
                },
                focusedRoundNumber: this.initialRoundNumber,
                leaderboards: this.populateInitialLeaderboard(),
                maxRoundNumber: this.initialRoundNumber,
                roundNumber: this.initialRoundNumber,
                teams,
                working: false,
            }
        },
        methods: {
            focusRound: function(roundNumber) {
                let component = this
                if (typeof component.leaderboards[roundNumber] != "undefined") {
                    component.focusedRoundNumber = roundNumber
                    return
                }
                component.resetErrorMessages()
                component.retrieveLeaderboard(roundNumber, true)
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
                this.roundNumber++
                this.focusRound(this.roundNumber)
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
                this.roundNumber--
                this.focusRound(this.roundNumber)
            },
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            retrieveLeaderboard: function(roundNumber, setFocus) {
                let component = this
                if (setFocus === true) {
                    component.working = true
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
                    component.errorMessages = ['Something went wrong :(<br />Please try again later']
                }).finally(function() {
                    component.working = false
                })
            },
            setLeaderboard: function(roundNumber, rankings, lastUpdatedUnix) {
                this.leaderboards = this.populateLeaderboardByRound(
                    this.leaderboards,
                    roundNumber,
                    rankings,
                    lastUpdatedUnix
                )
            },
            showScoredEntryPrediction: function(entryID, entryNickname, score) {
                this.focusedEntry.id = entryID
                this.focusedEntry.nickname = entryNickname
                this.focusedEntry.score = score
                $('#scoredEntryPredictionModal').modal('show')
            }
        },
        computed: {
            focusedLeaderboard: function() {
                return this.leaderboards[this.focusedRoundNumber]
            },
            lastUpdatedVerbose: function() {
                const helpers = require('../../helpers.js')
                if (this.focusedLeaderboard.lastUpdated === null) {
                    return ""
                }
                return helpers.formatVerboseDate(this.focusedLeaderboard.lastUpdated)
            },
        }
    }
</script>
