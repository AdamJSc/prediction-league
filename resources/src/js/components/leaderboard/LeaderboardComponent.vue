<template>
    <div class="leaderboard-container">
        <div id="scoredEntryPredictionModal" class="modal fade" tabindex="-1" role="dialog">
            <div class="modal-dialog" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">{{focusedEntry.nickname}}: Round {{roundNumber}}</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">
                        <scored-entry-selection v-bind:entry-id="focusedEntry.id" v-bind:round-number="roundNumber"></scored-entry-selection>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
                    </div>
                </div>
            </div>
        </div>
        <table class="round">
            <tr>
                <td><i class="fa fa-chevron-left" aria-hidden="true"></i></td>
                <td class="text-center">Round {{roundNumber}}</td>
                <td class="text-right"><i class="fa fa-chevron-right" aria-hidden="true"></i></td>
            </tr>
        </table>
        <div class="last-updated text-center">Last updated on {{lastUpdatedVerbose}}</div>
        <table class="leaderboard-render rankings">
            <thead>
            <tr>
                <td colspan="3"></td>
                <td class="text-right">
                    Pts
                </td>
                <td class="text-right">
                    Min
                </td>
                <td class="text-right">
                    Rnd
                </td>
            </tr>
            </thead>
            <tbody>
            <tr v-for="ranking in rankings" v-on:click="showScoredEntryPrediction(ranking.id, entries[ranking.id])" class="leaderboard-row rankings-row">
                <td class="position">
                    {{ranking.position}}
                </td>
                <td class="popout">
                    <i class="fa fa-external-link" aria-hidden="true"></i>
                </td>
                <td class="name">
                    {{entries[ranking.id]}}
                </td>
                <td class="text-right">
                    {{ranking.total_score}}
                </td>
                <td class="text-right">
                    {{ranking.min_score}}
                </td>
                <td class="text-right">
                    {{ranking.score}}
                </td>
            </tr>
            </tbody>
        </table>
    </div>
</template>

<script>
    const axios = require('axios').default

    export default {
        name: 'LeaderBoard',
        props: {
            roundNumber: {
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
            }
        },
        data: function() {
            let lastUpdated = null
            if (this.unix !== "0") {
                lastUpdated = new Date(parseFloat(this.unix + '000'))
            }

            let entries = this.rawEntries === "" ? [] : JSON.parse(this.rawEntries)
            let rankings = this.rawRankings === "" ? [] : JSON.parse(this.rawRankings)

            return {
                lastUpdated: lastUpdated,
                roundNumber: this.roundNumber,
                entries: entries,
                rankings: rankings,
                focusedEntry: {
                    id: "",
                    nickname: "",
                },
            }
        },
        mounted: function() {
        },
        methods: {
            showScoredEntryPrediction: function(entryID, entryNickname) {
                this.focusedEntry.id = entryID
                this.focusedEntry.nickname = entryNickname
                $('#scoredEntryPredictionModal').modal('toggle')
            }
        },
        computed: {
            lastUpdatedVerbose: function() {
                return this.lastUpdated.toISOString()
            },
        }
    }
</script>
