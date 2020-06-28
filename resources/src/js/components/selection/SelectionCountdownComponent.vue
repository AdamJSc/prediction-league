<template>
    <div class="selection-countdown-container">
        <p>{{label}} <span class="remaining">{{remainingVerbose}}</span></p>
    </div>
</template>

<script>
    export default {
        name: 'SelectionCountdown',
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
                case 'selection-close':
                    label = 'Current window closes in'
                    break
                case 'selection-open':
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
                let diff = this.target - this.now

                if (diff <= 0) {
                    return '0 seconds'
                }

                let inSeconds = Math.floor(diff / 1000)
                let inMinutes = Math.floor(diff / (1000 * 60))
                let inHours = Math.floor(diff / (1000 * 60 * 60))
                let inDays = Math.floor(diff / (1000 * 60 * 60 * 24))

                let numOfDays = inDays
                let numOfHours = inHours - (inDays * 24)
                let numOfMinutes = inMinutes - (inHours * 60)
                let numOfSeconds = inSeconds - (inMinutes * 60)

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
            setInterval(this.refreshNow, 1000)
        }
    }
</script>
