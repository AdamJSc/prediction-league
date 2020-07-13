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
        <div v-if="working" class="loader-container">
            <img alt="loader" src="/assets/img/loader-light-bg.svg" />
        </div>
        <div v-if="!working && rankings.length > 0" class="leaderboard-render-wrapper">
            <div class="last-updated text-center">Last updated on {{lastUpdatedVerbose}}</div>
            <table class="leaderboard-render rankings clickable">
                <thead>
                <tr>
                    <td colspan="3"></td>
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
                <tr v-for="ranking in rankings" v-on:click="showScoredEntryPrediction(ranking.id, entries[ranking.id], ranking.score)" class="leaderboard-row rankings-row">
                    <td class="position">
                        {{ranking.position}}
                    </td>
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
            seasonId: {
                type: String
            },
            initialRoundNumber: {
                type: Number
            },
            unix: {
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
        },
        data: function() {
            let lastUpdated = null
            if (this.unix !== "0") {
                lastUpdated = this.parseUnixDate(this.unix + '000')
            }

            let entries = this.rawEntries === "" ? [] : JSON.parse(this.rawEntries)
            let rankings = this.rawRankings === "" ? [] : JSON.parse(this.rawRankings)
            let teams = this.rawTeams === "" ? [] : JSON.parse(this.rawTeams)

            return {
                working: false,
                errorMessages: [],
                lastUpdated: lastUpdated,
                maxRoundNumber: this.initialRoundNumber,
                roundNumber: this.initialRoundNumber,
                entries: entries,
                rankings: rankings,
                teams: teams,
                focusedEntry: {
                    id: "",
                    nickname: "",
                    score: 0
                },
            }
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            parseUnixDate: function(timestamp) {
                return new Date(parseFloat(timestamp))
            },
            prevRound: function() {
                this.roundNumber--
                this.refreshLeaderboard()
            },
            nextRound: function() {
                this.roundNumber++
                this.refreshLeaderboard()
            },
            refreshLeaderboard: function() {
                const vm = this
                vm.working = true
                vm.resetErrorMessages()
                vm.lastUpdated = null
                vm.rankings = []

                const retrieveLeaderboard = function() {
                    let url = `/api/season/${vm.seasonId}/leaderboard/${vm.roundNumber}`
                    axios.request({
                        method: 'get',
                        url: url
                    })
                        .then(function (response) {
                            let unix = Date.parse(response.data.data.leaderboard.last_updated)
                            vm.lastUpdated = vm.parseUnixDate(unix)
                            vm.rankings = response.data.data.leaderboard.rankings
                            vm.working = false
                        })
                        .catch(function (error) {
                            console.log(error)
                            vm.errorMessages = ['Something went wrong :(<br />Please try again later']
                            vm.working = false
                        })
                }

                setTimeout(retrieveLeaderboard, 500)
            },
            showScoredEntryPrediction: function(entryID, entryNickname, score) {
                this.focusedEntry.id = entryID
                this.focusedEntry.nickname = entryNickname
                this.focusedEntry.score = score
                $('#scoredEntryPredictionModal').modal('show')
            }
        },
        computed: {
            lastUpdatedVerbose: function() {
                const helpers = require('../../helpers.js')
                return helpers.formatVerboseDate(this.lastUpdated)
            },
        }
    }
</script>
