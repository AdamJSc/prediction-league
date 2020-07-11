<template>
    <div class="scored-entry-prediction-container">
        <transition name="fade">
            <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                <p v-for="msg in errorMessages" v-html="msg"></p>
            </div>
        </transition>
        <div v-if="working" class="loader-container">
            <img alt="loader" src="/assets/img/loader-light-bg.svg" />
        </div>
        <div v-else>
            <table>
                <tr>
                    <td colspan="4">Round Score = {{roundScore}}</td>
                </tr>
                <tr v-for="ranking in rankings">
                    <td>{{ranking.position}}</td>
                    <td>{{ranking.id}}</td>
                    <td>{{ranking.score}}</td>
                    <td>{{ranking.meta_position}}</td>
                </tr>
            </table>
        </div>
    </div>
</template>

<script>
    const axios = require('axios').default

    export default {
        name: 'ScoredEntryPrediction',
        props: {
            entryId: {
                type: String
            },
            roundNumber: {
                type: Number
            }
        },
        data: function() {
            return {
                working: false,
                errorMessages: [],
                roundScore: 0,
                rankings: []
            }
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            refreshScoredEntryPrediction: function(entryId) {
                const vm = this
                vm.working = true
                vm.roundScore = 0
                vm.rankings = []

                let url = `/api/entry/${entryId}/scored/${vm.roundNumber}`
                axios.request({
                    method: 'get',
                    url: url
                })
                    .then(function (response) {
                        vm.roundScore = response.data.data.scored.round_score
                        vm.rankings = response.data.data.scored.rankings
                        vm.working = false
                    })
                    .catch(function (error) {
                        console.log(error)
                        vm.errorMessages = ['Something went wrong :(<br />Please try again later']
                        vm.working = false
                    })
            }
        },
        watch: {
            entryId: function(newEntryId) {
                this.refreshScoredEntryPrediction(newEntryId)
            }
        }
    }
</script>
