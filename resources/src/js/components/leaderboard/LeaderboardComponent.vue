<template>
  <div class="leaderboard-container">
    <round-navigation
        :is-working="isWorking"
        :max-round-number="maxRoundNumber"
        :round-number="roundNumber"
    ></round-navigation>
    <transition name="fade">
      <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
        <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
        <p v-for="msg in errorMessages" v-html="msg"></p>
      </div>
    </transition>
    <div v-if="leaderboardToShow.rankings.length > 0" class="leaderboard-render-wrapper">
      <div v-if="isWorking" class="loader-container">
        <img alt="loader" src="/assets/img/loader-light-bg.svg" />
      </div>
      <div v-if="!isWorking && leaderboardToShow.lastUpdated" class="last-updated text-center">Updated {{lastUpdatedVerbose}}</div>
      <table class="leaderboard-render rankings clickable">
        <thead>
        <tr>
          <td colspan="3"></td>
          <td class="text-right text-lolight">
            Score
          </td>
          <td class="text-right text-highlight">
            Total
          </td>
        </tr>
        </thead>
        <tbody>
        <tr v-for="entry in leaderboardToShow.rankings" v-on:click="$emit('entry-change', entry.id)" class="leaderboard-row rankings-row">
          <td class="position">
            {{entry.position}}
          </td>
          <td class="movement" v-html="getMovementMarkup(entry.movement)"></td>
          <td class="name text-highlight">
            {{entries[entry.id]}} <i class="fa-solid fa-arrow-up-right-from-square"></i>
          </td>
          <td class="text-right text-lolight">
            {{entry.score}}
          </td>
          <td class="text-right text-highlight">
            {{entry.total_score}}
          </td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
  const axios = require('axios').default
  const preloadBuffer = 2

  export default {
    name: 'Leaderboard',
    props: {
      entries: { // map of entry id to entry nickname
        type: Object
      },
      initialLastUpdatedUnix: { // last updated timestamp of leaderboard relating to initial round number
        type: String
      },
      initialRankings: { // array of rankings objects that pertain to initial round number, with the schema id, position, score, min_score, total_score, movement
        type: Array
      },
      roundNumber: { // round number that is inc/decremented by navigation controls
        type: Number
      },
      seasonId: { // id of the season that is being viewed
        type: String
      },
      teams: { // array of team objects with the schema id, name, short_name, crest_url
        type: Array
      }
    },
    data: function() {
      return {
        isWorking: false, // denotes whether leaderboards are in the process of being retrieved
        errorMessages: [], // error messages relating to retrieval of leaderboard
        leaderboards: this.applyLeaderboardRankings( // map of leaderboards indexed by round number
            {},
            this.roundNumber,
            this.initialRankings,
            parseInt(this.initialLastUpdatedUnix + '000')
        ),
        maxRoundNumber: this.roundNumber, // maximum available round number
        showRoundNumber: this.roundNumber, // the round number to display
      }
    },
    methods: {
      applyLeaderboardRankings: function(leaderboards, roundNumber, rankings, lastUpdatedUnix) {
        let lastUpdated = new Date(lastUpdatedUnix)
        leaderboards[roundNumber] = {rankings, lastUpdated}
        return leaderboards
      },
      getLeaderboardToShow: function(roundNumber) {
        return this.leaderboardExists(roundNumber) ? this.leaderboards[roundNumber] : []
      },
      getMovementMarkup: function(movement) {
        if (movement > 0) {
          return '<span class="movement-up"><i class="fas fa-caret-up"/></span>'
        }
        if (movement < 0) {
          return '<span class="movement-down"><i class="fas fa-caret-down"/></span>'
        }
        return '<span class="movement-none"><i class="fas fa-minus"></i></span>'
      },
      leaderboardExists: function(roundNumber) {
        return typeof this.leaderboards[roundNumber] !== 'undefined'
      },
      resetErrorMessages: function() {
        this.errorMessages = []
      },
      retrieveLeaderboard: function(roundNumber, isForeground) {
        const component = this

        const showIfForeground = function() {
          if (isForeground) {
            component.showRoundNumber = roundNumber
          }
        }

        if (roundNumber < 1 || roundNumber > this.maxRoundNumber) {
          return
        }
        if (this.leaderboardExists(roundNumber)) {
          showIfForeground()
          return
        }

        if (isForeground) {
          component.isWorking = true
        }

        axios.request({
          method: 'get',
          url: `/api/season/${component.seasonId}/leaderboard/${roundNumber}`
        }).then(function(response) {
          let data = response.data.data
          let rankings = data.leaderboard.rankings
          let lastUpdatedUnix = Date.parse(data.leaderboard.last_updated)
          component.setLeaderboardRankings(roundNumber, rankings, lastUpdatedUnix)
          showIfForeground()
        }).catch(function(error) {
          if (isForeground) {
            console.log(error)
            component.errorMessages = ['Something went wrong :(<br />Please try again later']
          }
        }).finally(function() {
          component.isWorking = false
        })
      },
      retrieveLeaderboards: function(lowerRound, upperRound, foregroundRound) {
        for (let roundNumber = lowerRound; roundNumber <= upperRound; roundNumber++) {
          let isForeground = (roundNumber === foregroundRound)
          this.retrieveLeaderboard(roundNumber, isForeground)
        }
      },
      setLeaderboardRankings: function(roundNumber, rankings, lastUpdatedUnix) {
        this.leaderboards = this.applyLeaderboardRankings(this.leaderboards, roundNumber, rankings, lastUpdatedUnix)
      },
    },
    computed: {
      lastUpdatedVerbose: function() {
        const helpers = require('../../helpers.js')
        let leaderboard = this.getLeaderboardToShow(this.showRoundNumber)
        if (leaderboard.lastUpdated === null) {
          return ""
        }
        return helpers.formatVerboseDate(leaderboard.lastUpdated)
      },
      leaderboardToShow: function() {
        return this.getLeaderboardToShow(this.showRoundNumber)
      }
    },
    watch: {
      roundNumber: function(newRoundNumber) {
        this.resetErrorMessages()
        let lower = newRoundNumber - preloadBuffer
        let upper = newRoundNumber + preloadBuffer
        this.retrieveLeaderboards(lower, upper, newRoundNumber)
      },
    },
    mounted() {
      // preload initial number of "buffer" leaderboards
      let lower = this.maxRoundNumber - preloadBuffer
      this.retrieveLeaderboards(lower, this.maxRoundNumber, this.roundNumber)
    }
  }
</script>
