<template>
    <div class="col-md-8 offset-md-2">
        Teams = {{teamsCount}}
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
        },
        computed: {
            teamsCount: function() {
                return this.teams.length
            }
        },
        mounted: function() {
            let teams = this.teams
            setInterval(function() {
                teams.pop()
            }, 1000)
        }
    }
</script>
