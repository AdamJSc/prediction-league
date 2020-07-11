<template>
    <p>{{entryId}}: {{roundNumber}}</p>
</template>

<script>
    export default {
        name: 'ScoredEntrySelection',
        props: {
            entryId: {
                type: String
            },
            roundNumber: {
                type: Number
            }
        },
        watch: {
            entryId: function(newEntryId) {
                console.log('entryId has changed to ' + newEntryId)
            }
        }
    }
</script>
