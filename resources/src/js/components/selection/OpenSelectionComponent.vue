<template>
    <div class="col-md-8 offset-md-2">
        <div class="teams-container">
            <p>Click & Drag to reorder - don't forget to click "Save" at the bottom once you're done!</p>
            <div v-for="(id, index) in teamsIDSequence" v-on:click="teamOnClick" v-bind:team-id="id"
                 v-bind:class="['team-wrapper', { 'selected': isSelected(id) }]">
                <div class="dragger"><i class="fa fa-bars" aria-hidden="true"></i></div>
                <div class="position">{{index+1}}</div>
                <div class="crest"><img :alt="getTeamByID(id).name" :src="getTeamByID(id).crest_url" /></div>
                <div class="name">{{getTeamByID(id).short_name}}</div>
            </div>
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
                selectedTeamID: null
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
                return this.selectedTeamID === id
            },
            swapTeams: function(idA, idB) {
                let indexOfA = this.teamsIDSequence.indexOf(idA)
                let indexOfB = this.teamsIDSequence.indexOf(idB)

                this.teamsIDSequence[indexOfA] = idB
                this.teamsIDSequence[indexOfB] = idA
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
