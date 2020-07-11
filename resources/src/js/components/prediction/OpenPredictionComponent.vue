<template>
    <div class="teams-container">
        <p v-if="dirtyTeamIDs.length > 0">Changed {{dirtyTeamIDs.length}} team(s).</p>
        <p v-if="!isSelected()">Select a team to change their position.</p>
        <p v-else>Who do you want to swap {{getTeamByID(selectedTeamID).short_name}} with?</p>
        <table class="teams-reorder rankings">
            <tbody>
            <tr v-for="(id, index) in teamsIDSequence" v-on:click="teamOnClick" v-bind:team-id="id"
                v-bind:class="['team-row', 'rankings-row', { 'selected': isSelected(id), 'dirty': isDirty(id) && !isSelected(id) }]">
                <td class="position">{{index+1}}</td>
                <td class="crest-outer"><div class="crest"><img :alt="getTeamByID(id).name" :src="getTeamByID(id).crest_url" /></div></td>
                <td><span class="name">{{getTeamByID(id).short_name}}</span></td>
            </tr>
            </tbody>
        </table>
        <table class="teams-reorder-admin">
            <tbody>
                <tr>
                    <td colspan="3">
                        <div class="call-to-action-wrapper">
                            <div v-if="lastUpdated" class="text-center">
                                Last updated on {{lastUpdatedVerbose}}
                            </div>
                            <div class="submit-wrapper">
                                <transition name="fade">
                                    <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                                        <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                                        <p v-for="msg in errorMessages" v-html="msg"></p>
                                    </div>
                                </transition>
                                <div v-if="success" class="alert alert-block alert-success">
                                    Your prediction has been updated. Good luck!
                                </div>
                                <div v-else>
                                    <action-button
                                            label="Update"
                                            @clicked="updateOnClick"
                                            :is-disabled="dirtyTeamIDs.length === 0 || working"
                                            :is-working="working"></action-button>
                                    <button v-on:click="resetState" class="btn btn-secondary">Reset</button>
                                </div>
                            </div>
                        </div>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<script>
    const axios = require('axios').default

    export default {
        name: 'OpenPrediction',
        props: {
            entry: {
                type: Object,
            },
            rawTeams: {
                type: String,
            },
            unix: {
                type: String
            }
        },
        data: function() {
            let lastUpdated = null
            if (this.unix !== "0") {
                lastUpdated = new Date(parseFloat(this.unix + '000'))
            }

            return {
                working: false,
                success: false,
                errorMessages: [],
                teams: JSON.parse(this.rawTeams),
                lastUpdated: lastUpdated,
                teamsIDSequence: [],
                selectedTeamID: null,
                dirtyTeamIDs: []
            }
        },
        mounted: function() {
            this.resetState()
        },
        methods: {
            getTeamByID: function(id) {
                for (let i = 0; i < this.teams.length; i++) {
                    if (this.teams[i].id === id) {
                        return this.teams[i]
                    }
                }
            },
            isSelected: function(id) {
                if (typeof id === 'undefined') {
                    return this.selectedTeamID !== null
                }
                return this.selectedTeamID === id
            },
            isDirty: function(id) {
                return this.dirtyTeamIDs.indexOf(id) !== -1
            },
            swapTeams: function(idA, idB) {
                let indexOfA = this.teamsIDSequence.indexOf(idA)
                let indexOfB = this.teamsIDSequence.indexOf(idB)

                // swap teams
                this.teamsIDSequence[indexOfA] = idB
                this.teamsIDSequence[indexOfB] = idA

                // refresh dirty teams
                this.dirtyTeamIDs = []
                for (let i = 0; i < this.teamsIDSequence.length; i++) {
                    if (this.teamsIDSequence[i] !== this.teams[i].id) {
                        this.dirtyTeamIDs.push(this.teamsIDSequence[i])
                    }
                }
            },
            resetState: function() {
                this.teamsIDSequence = []
                this.dirtyTeamIDs = []
                this.selectedTeamID = null

                for (let i = 0; i < this.teams.length; i++) {
                    this.teamsIDSequence.push(this.teams[i].id)
                }
            },
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            storeReorderedTeams: function() {
                let teams = []
                for (let i = 0; i < this.teamsIDSequence.length; i++) {
                    let id = this.teamsIDSequence[i]
                    teams.push(this.getTeamByID(id))
                }
                this.teams = teams
            },
            teamOnClick: function(e) {
                if (this.success) {
                    return
                }

                let id = e.currentTarget.getAttribute('team-id')
                switch (this.selectedTeamID) {
                    case id:
                        // team has been de-selected
                        this.selectedTeamID = null
                        return
                    case null:
                        // team has been selected
                        this.selectedTeamID = id
                        return
                }
                // swap this team with the already-selected team and reset selected team ID
                this.swapTeams(this.selectedTeamID, id)
                this.selectedTeamID = null
            },
            updateOnClick: function(e) {
                const vm = this
                vm.working = true

                let url = `/api/entry/${vm.entry.id}/prediction`
                axios.request({
                    method: 'post',
                    url: url,
                    data: {
                        entry_short_code: vm.entry.shortCode,
                        ranking_ids: vm.teamsIDSequence,
                    }
                })
                    .then(function (response) {
                        vm.working = false
                        vm.success = true
                        vm.lastUpdated = new Date()
                        vm.storeReorderedTeams()
                        vm.resetState()
                    })
                    .catch(function (error) {
                        console.log(error)
                        vm.errorMessages = ['Something went wrong :(<br />Please try again later']
                        vm.working = false
                    })
            }
        },
        computed: {
            lastUpdatedVerbose: function() {
                return this.lastUpdated.toISOString()
            },
        }
    }
</script>
