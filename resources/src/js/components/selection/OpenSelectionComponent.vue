<template>
    <div class="col-md-8 offset-md-2">
        <div class="teams-container">
            <p v-if="dirtyTeamIDs.length > 0">Changed {{dirtyTeamIDs.length}} team(s).</p>
            <p v-if="!isSelected()">Select a team to change their position.</p>
            <p v-else>Who do you want to swap {{getTeamByID(selectedTeamID).short_name}} with?</p>
            <table class="teams-reorder">
                <tbody>
                    <tr v-for="(id, index) in teamsIDSequence" v-on:click="teamOnClick" v-bind:team-id="id"
                        v-bind:class="['team-row', { 'selected': isSelected(id), 'dirty': isDirtyAndNotSelected(id) }]">
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
                                    Last updated on {{lastUpdated.format('ddd Do MMM [at] h:mma')}}
                                </div>
                                <div class="submit-wrapper">
                                    <button v-on:click="updateOnClick" class="btn btn-primary" v-bind:disabled="dirtyTeamIDs.length === 0 || working">
                                        <span v-if="working">Working...</span>
                                        <span v-else>Update</span>
                                    </button>
                                    <button v-on:click="resetState" class="btn btn-secondary">Reset</button>
                                </div>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<script>
    const moment = require('moment')

    export default {
        name: 'OpenSelection',
        props: {
            teamsRaw: {
                type: String,
            },
            unix: {
                type: String
            }
        },
        data: function() {
            let lastUpdated = null
            if (this.unix !== "0") {
                lastUpdated = moment.unix(this.unix)
            }

            return {
                working: false,
                teams: JSON.parse(this.teamsRaw),
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
            isDirtyAndNotSelected: function(id) {
                if (this.isSelected(id)) {
                    return false
                }
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
            storeReorderedTeams: function() {
                let teams = []
                for (let i = 0; i < this.teamsIDSequence.length; i++) {
                    let id = this.teamsIDSequence[i]
                    teams.push(this.getTeamByID(id))
                }
                this.teams = teams
            },
            teamOnClick: function(e) {
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
                this.working = true
                this.lastUpdated = moment()
                this.storeReorderedTeams()
                this.resetState()
            }
        },
    }
</script>
