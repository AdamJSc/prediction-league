<template>
    <div class="prediction-countdown-container">
        <p>{{label}} <span class="remaining">{{remainingVerbose}}</span></p>
    </div>
</template>

<script>
    export default {
        name: 'PredictionCountdown',
        props: {
            unix: {
                type: String,
            },
            subject: {
                type: String,
            }
        },
        data: function() {
            let label = ''
            switch (this.subject) {
                case 'prediction-close':
                    label = 'Current window closes in'
                    break
                case 'prediction-open':
                    label = 'Next window opens in'
                    break
            }
            return {
                now: new Date(),
                target: new Date(parseFloat(this.unix + '000')),
                label: label,
            }
        },
        methods: {
            refreshNow: function() {
                this.now = new Date()
            }
        },
        computed: {
            remainingVerbose: function() {
                const helpers = require('../../helpers.js')
                return helpers.formatVerboseDuration(this.now, this.target)
            }
        },
        mounted: function() {
            setInterval(this.refreshNow, 1000)
        }
    }
</script>
