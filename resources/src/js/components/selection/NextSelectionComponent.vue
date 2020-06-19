<template>
    <div class="col-md-8 offset-md-2">
        <p>Selections Window opens in:</p>
        <p>{{remainingVerbose}}</p>
    </div>
</template>

<script>
    const moment = require('moment')

    export default {
        name: 'NextSelection',
        props: {
            unix: {
                type: String,
            }
        },
        data: function() {
            return {
                remaining: this.getRemaining(this.unix)
            }
        },
        methods: {
            getRemaining: function(unix) {
                let now = moment()
                let target = moment.unix(unix)

                return moment.duration(target.diff(now))
            },
            decrementRemaining: function() {
                this.remaining.subtract(1, 's')
            }
        },
        computed: {
            remainingVerbose: function() {
                let remaining = this.remaining

                if (remaining.asSeconds() < 0) {
                    return '0 seconds'
                }

                let numOfDays = remaining.days()
                let numOfHours = remaining.hours()
                let numOfMinutes = remaining.minutes()
                let numOfSeconds = remaining.seconds()

                let days = '', hours = '', minutes = '', seconds = ''

                if (numOfDays > 0) {
                    days = `${numOfDays} day` + (numOfDays !== 1 ? 's' : '')
                }

                if (numOfDays > 0 || numOfHours > 0) {
                    hours = `${numOfHours} hour` + (numOfHours !== 1 ? 's' : '')
                }

                if (numOfDays > 0 || numOfHours > 0 || numOfMinutes > 0) {
                    minutes = `${numOfMinutes} minute` + (numOfMinutes !== 1 ? 's' : '')
                }

                seconds = `${numOfSeconds} second` + (numOfSeconds !== 1 ? 's' : '')

                let verbose = `${days} ${hours} ${minutes} ${seconds}`

                return verbose.trim()
            }
        },
        mounted: function() {
            setInterval(this.decrementRemaining, 1000)
        }
    }
</script>
