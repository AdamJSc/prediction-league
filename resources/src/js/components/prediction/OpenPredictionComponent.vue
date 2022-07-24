<template>
    <div class="teams-container">
        <div v-if="canReorder" class="text-center">
          <p><strong>Make your changes!</strong></p>
          <p>Select two teams to swap their positions.</p>
        </div>
        <div v-else class="text-center alert alert-block alert-info">
            You have already made the maximum number of permitted changes for the current Match Week.
        </div>
        <table v-bind:class="['teams-reorder', 'rankings', { 'clickable': canReorder }]">
          <tbody>
            <tr v-for="(id, index) in teamsIDSequence" v-on:click="teamOnClick" v-bind:team-id="id"
                v-bind:class="['team-row', 'rankings-row', { 'selected': isSelected(id), 'dirty': isDirty(id) && !isSelected(id) }]">
              <td class="position">{{index+1}}</td>
              <td class="crest-outer" style="height:100%">
                <div class="bg-img-fill" :style="getTeamCrestBGStyle(id)"></div>
              </td>
              <td>
                <span class="name">{{getTeamByID(id).short_name}}</span>
              </td>
            </tr>
          </tbody>
        </table>
        <table class="teams-reorder-admin">
            <tbody>
                <tr>
                    <td colspan="3">
                        <div class="call-to-action-wrapper">
                            <div v-if="lastUpdated" class="text-center">
                                Updated {{lastUpdatedVerbose}}
                            </div>
                            <div class="submit-wrapper">
                                <transition name="fade">
                                    <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                                        <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                                        <p v-for="msg in errorMessages" v-html="msg"></p>
                                    </div>
                                </transition>
                                <div v-if="dirtyTeamIDs.length > 0" class="alert alert-warning">
                                    Changed {{dirtyTeamIDs.length}} team(s).
                                </div>
                                <div v-if="success" class="alert alert-block alert-success">
                                    Your table has been updated. Good luck!
                                </div>
                                <div v-else-if="canReorder && exceedsLimit" class="alert alert-block alert-danger">
                                    Sorry, you can only change up to {{predLimit}} teams! <a v-on:click="resetState" href="#">[Reset]</a>
                                </div>
                                <div v-else-if="canReorder">
                                    <action-button
                                            label="Update"
                                            @clicked="updateOnClick"
                                            :is-disabled="!canUpdate"
                                            :is-working="working"
                                            :is-primary="true"></action-button>
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
                type: Object
            },
            predLimit: {
                type: Number
            },
            rawTeams: {
                type: String
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
            getTeamCrestBGStyle: function(id) {
                let team = this.getTeamByID(id)
                return `background-image:url('${team.crest_url}')`
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
                if (this.success || !this.canReorder) {
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
                        entry_pred_token: vm.entry.predToken,
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
                const helpers = require('../../helpers.js')
                return helpers.formatVerboseDate(this.lastUpdated)
            },
            canUpdate: function() {
                if (this.working) {
                    // update is already in process
                    return false
                }
                if (this.dirtyTeamIDs.length === 0) {
                    // no teams have been changed
                    return false
                }
                if (this.exceedsLimit) {
                    // limit exceeded
                    return false
                }
                // otherwise, ok to update
                return true
            },
            canReorder: function() {
                // can reorder teams if limit is anything other than 0
                return !(this.predLimit === 0)
            },
            exceedsLimit: function() {
                if (this.predLimit === -1) {
                    // unlimited teams can be changed
                    return false
                }
                if (this.predLimit === 0) {
                    // no teams can be changed
                    return true
                }
                if (this.dirtyTeamIDs.length > this.predLimit) {
                    // number of changed teams exceeds the permitted limit
                    return true
                }
                return false
            }
        }
    }
</script>
