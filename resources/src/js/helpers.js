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

    return this
}

module.exports = helpers()