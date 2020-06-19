<template>
    <div class="col-md-8 offset-md-2">
        <div class="teams-container">
            <p>Drag to reorder</p>
            <div v-for="(team, index) in teams" class="team-wrapper">
                <div class="dragger"><i class="fa fa-bars" aria-hidden="true"></i></div>
                <div class="position">{{index+1}}</div>
                <div class="crest"><img :alt="team.name" :src="team.crest_url" /></div>
                <div class="name">{{team.short_name}}</div>
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
                teams: JSON.parse(this.teamsPayload)
            }
        }
    }
</script>
