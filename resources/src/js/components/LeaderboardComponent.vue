<template>
    <div class="leaderboard-container">
        <p>Round {{roundNumber}}</p>
        <p>Last updated on {{lastUpdatedVerbose}}</p>
        <table class="leaderboard-render">
            <thead>
            <tr>
                <td colspan="2"></td>
                <td>
                    <p>Score</p>
                </td>
                <td>
                    <p>Min</p>
                </td>
                <td>
                    <p>Pts</p>
                </td>
            </tr>
            </thead>
            <tbody>
            <tr v-for="ranking in rankings">
                <td>
                    <p>{{ranking.position}}</p>
                </td>
                <td>
                    <p>{{entries[ranking.id]}}</p>
                </td>
                <td>
                    <p>{{ranking.score}}</p>
                </td>
                <td>
                    <p>{{ranking.min_score}}</p>
                </td>
                <td>
                    <p>{{ranking.total_score}}</p>
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
