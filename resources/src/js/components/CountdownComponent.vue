<template>
    <div v-if="shouldShow" class="countdown-container">
        <div class="countdown highlight-container">
            <p>{{label}}</p>
            <p class="remaining-time">{{remainingVerbose}}</p>
        </div>
    </div>
</template>

<script>
    export default {
        name: 'CountdownComponent',
        props: {
            unix: {
                type: String
            },
            label: {
                type: String
            }
        },
        data: function() {
            return {
                target: new Date(parseFloat(this.unix + '000')),
                now: new Date()
            }
        },
        mounted: function() {
            setInterval(this.refreshNow, 1000)
        },
        methods: {
            refreshNow: function() {
                this.now = new Date()
            }
        },
        computed: {
            shouldShow: function() {
                let oneHour = 3600000
                // within 24 hours
                return (this.target - this.now) < (24 * oneHour)
            },
            remainingVerbose: function() {
                const helpers = require('../helpers.js')
                return helpers.formatVerboseDuration(this.now, this.target)
            }
        }
    }
</script>