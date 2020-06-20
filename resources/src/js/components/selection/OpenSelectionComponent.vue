<template>
    <div class="col-md-8 offset-md-2">
        <div class="teams-container">
            <p v-if="!isSelected()">Select a team to change their position.</p>
            <p v-else>Who do you want to swap {{getTeamByID(selectedTeamID).short_name}} with?</p>
            <table class="teams-table">
                <tbody>
                <tr v-for="(id, index) in teamsIDSequence" v-on:click="teamOnClick" v-bind:team-id="id"
                    v-bind:class="['team-row', { 'selected': isSelected(id), 'dirty': isDirtyAndNotSelected(id) }]">
                    <td class="position">{{index+1}}</td>
                    <td class="crest-outer"><div class="crest"><img :alt="getTeamByID(id).name" :src="getTeamByID(id).crest_url" /></div></td>
                    <td><span class="name">{{getTeamByID(id).short_name}}</span></td>
                </tr>
                </tbody>
            </table>
            <button class="btn btn-primary">Save</button>
        </div>
    </div>
</template>

<script>
    export default {
        name: 'OpenSelection',
        props: {
            teamsPayload: {
                type: String,
            }
        },
        data: function() {
            return {
                teams: JSON.parse(this.teamsPayload),
                teamsIDSequence: [],
                selectedTeamID: null,
                dirtyTeamIDs: []
            }
        },
        mounted: function() {
            for (let i = 0; i < this.teams.length; i++) {
                this.teamsIDSequence.push(this.teams[i].id)
            }
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
            }
        },
    }
</script>
