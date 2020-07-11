<template>
    <div class="leaderboard-container">
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
                <td colspan="2"></td>
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
            <tr v-for="ranking in rankings" class="leaderboard-row rankings-row">
                <td class="position">
                    {{ranking.position}}
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
            }
        },
        mounted: function() {
        },
        methods: {
        },
        computed: {
            lastUpdatedVerbose: function() {
                return this.lastUpdated.toISOString()
            },
        }
    }
</script>
