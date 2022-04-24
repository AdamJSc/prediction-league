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
    <div v-if="leaderboardForRound.rankings.length > 0" class="leaderboard-render-wrapper">
      <div v-if="isWorking" class="loader-container">
        <img alt="loader" src="/assets/img/loader-light-bg.svg" />
      </div>
      <div v-if="!isWorking && leaderboardForRound.lastUpdated" class="last-updated text-center">Last updated on {{lastUpdatedVerbose}}</div>
      <table class="leaderboard-render rankings clickable">
        <thead>
        <tr>
          <td colspan="4"></td>
          <td class="text-right text-highlight">
            Pts
          </td>
          <td class="text-right text-lolight">
            Rnd
          </td>
          <td class="text-right text-lolight">
            Min
          </td>
        </tr>
        </thead>
        <tbody>
        <tr v-for="ranking in leaderboardForRound.rankings" v-on:click="$emit('entry-change', ranking.id)" class="leaderboard-row rankings-row">
          <td class="position">
            {{ranking.position}}
          </td>
          <td class="movement" v-html="getMovementMarkup(ranking.movement)"></td>
          <td class="popout text-highlight">
            <i class="fa fa-external-link" aria-hidden="true"></i>
          </td>
          <td class="name text-highlight">
            {{entries[ranking.id]}}
          </td>
          <td class="text-right text-highlight">
            {{ranking.total_score}}
          </td>
          <td class="text-right text-lolight">
            {{ranking.score}}
          </td>
          <td class="text-right text-lolight">
            {{ranking.min_score}}
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
      initialRankings: { // array of rankings objects that pertain to initial round number
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
            this.initialLastUpdatedUnix
        ),
        roundNumber: {},
        maxRoundNumber: this.roundNumber, // maximum available round number
      }
    },
    methods: {
      applyLeaderboardRankings: function(leaderboards, roundNumber, rankings, lastUpdatedUnix) {
        leaderboards[roundNumber] = {rankings, lastUpdatedUnix}
        return leaderboards
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
        if (roundNumber < 1 || roundNumber > this.maxRoundNumber) {
          return
        }
        if (this.leaderboardExists(roundNumber)) {
          return
        }

        if (isForeground) {
          this.working = true
        }

        const component = this

        axios.request({
          method: 'get',
          url: `/api/season/${component.seasonId}/leaderboard/${roundNumber}`
        }).then(function(response) {
          let data = response.data.data
          let rankings = data.leaderboard.rankings
          let lastUpdatedUnix = Date.parse(data.leaderboard.last_updated)
          component.setLeaderboardRankings(roundNumber, rankings, lastUpdatedUnix)
        }).catch(function(error) {
          console.log(error)
          if (isForeground) {
            component.errorMessages = ['Something went wrong :(<br />Please try again later']
          }
        }).finally(function() {
          component.working = false
        })
      },
      retrieveLeaderboards: function(lower, upper, foreground) {
        for (let roundNumber = lower; roundNumber <= upper; roundNumber++) {
          let isForeground = (roundNumber === foreground)
          this.retrieveLeaderboard(roundNumber, isForeground)
        }
      },
      setLeaderboardRankings: function(roundNumber, rankings, lastUpdatedUnix) {
        this.leaderboards = this.applyLeaderboardRankings(this.leaderboards, roundNumber, rankings, lastUpdatedUnix)
      },
    },
    computed: {
      leaderboardForRound: function() {
        return this.leaderboardExists(this.roundNumber) ? this.leaderboards[this.roundNumber] : null
      },
      lastUpdatedVerbose: function() {
        const helpers = require('../../helpers.js')
        if (this.leaderboardForRound.lastUpdated === null) {
          return ""
        }
        return helpers.formatVerboseDate(this.leaderboardForRound.lastUpdated)
      },
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
      let lower = this.maxRoundNumber-preloadBuffer
      this.retrieveLeaderboards(lower, this.maxRoundNumber)
    }
  }
</script>
