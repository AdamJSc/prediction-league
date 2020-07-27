const helpers = () => {
    this.formatVerboseDate = (dateToFormat) => {
        const months = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December']
        const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

        if (typeof dateToFormat === null) {
            return ""
        }

        let year = dateToFormat.getFullYear()
        let month = dateToFormat.getMonth()
        let date = dateToFormat.getDate()
        let hour = dateToFormat.getHours()
        let min = dateToFormat.getMinutes()
        let weekday = dateToFormat.getDay()
        let ampm = 'am'

        if(hour >= 12) {
            ampm = 'pm'
            hour = hour - 12
        }
        if (hour === 0) {
            hour = 12
        }

        if (min < 10) {
            min = '0' + min
        }

        return `${days[weekday]} ${date} ${months[month]} ${year} at ${hour}:${min}${ampm}`
    }

    this.formatVerboseDuration = (start, target) => {
        let diff = target - start

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

    return this
}

module.exports = helpers()