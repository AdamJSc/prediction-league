<template>
  <div class="scored-entry-container">
    <div class="text-center">
      <h3>{{scoredEntryToShow.entryNickname}}</h3>
    </div>
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
    <div v-if="isWorking" class="loader-container">
      <img alt="loader" src="/assets/img/loader-light-bg.svg" />
    </div>
    <div v-if="!isWorking && errorMessages.length == 0">
      <div class="scores-container text-center">
        <div class="round-score-container">
          <div class="subheading">Score</div>
          <div class="score">{{scoredEntryToShow.roundScore}}</div>
        </div>
      </div>
      <table class="rankings full-width">
        <thead>
        <tr>
          <td colspan="3"></td>
          <td class="text-right text-highlight">Pts</td>
          <td class="text-right text-lolight">Pos</td>
        </tr>
        </thead>
        <tbody>
        <tr v-for="ranking in scoredEntryToShow.rankings" class="rankings-row">
          <td class="position">{{ranking.position}}</td>
          <td class="crest-outer"><div class="crest"><img :alt="getTeamByID(ranking.id).name" :src="getTeamByID(ranking.id).crest_url" /></div></td>
          <td><span class="name">{{getTeamByID(ranking.id).short_name}}</span></td>
          <td class="score text-right text-highlight">{{ranking.score}}</td>
          <td class="meta_position text-right text-lolight">{{ranking.meta_position}}</td>
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
    name: 'ScoredEntry',
    props: {
      entryId: { // entry id to retrieve scored entries on behalf of
        type: String
      },
      entryNickname: { // nickname of the provided entry id
        type: String
      },
      roundNumber: { // round number that is inc/decremented by navigation controls
        type: Number
      },
      teams: { // array of team objects with the schema id, name, short_name, crest_url
        type: Object
      }
    },
    data: function() {
      return {
        isWorking: false, // denotes whether scored entries are in the process of being retrieved
        errorMessages: [], // error messages relating to retrieval of scored entries
        maxRoundNumber: this.roundNumber, // maximum available round number
        scoredEntries: {}, // map of scored entries indexed by entry id and round number
        showRoundNumber: 0, // the round number to display
      }
    },
    methods: {
      getScoredEntryToShow: function(entryId, roundNumber) {
        return this.scoredEntryExists(entryId, roundNumber) ? this.scoredEntries[entryId][roundNumber] : {}
      },
      getTeamByID: function(id) {
        for (let i in this.teams) {
          if (this.teams[i].id === id) {
            return this.teams[i]
          }
        }
        return null
      },
      resetErrorMessages: function() {
        this.errorMessages = []
      },
      retrieveScoredEntry: function(entryId, roundNumber, isForeground) {
        if (entryId == null || entryId === '') {
          return
        }

        const component = this

        const showIfForeground = function() {
          if (isForeground) {
            component.showRoundNumber = roundNumber
          }
        }

        if (roundNumber < 1 || roundNumber > this.maxRoundNumber) {
          return
        }
        if (this.scoredEntryExists(entryId, roundNumber)) {
          showIfForeground()
          return
        }

        if (isForeground) {
          component.isWorking = true
        }

        axios.request({
          method: 'get',
          url: `/api/entry/${entryId}/scored/${roundNumber}`
        }).then(function(response) {
          // get array of rankings objects that pertain to initial round number, with the schema id, position, score, max_score, total_score, movement
          let rankings = response.data.data.scored.rankings
          let roundScore = response.data.data.scored.round_score
          for (let i in rankings) {
            if (component.getTeamByID(rankings[i].id) === null) {
              if (isForeground) {
                console.log(`team with id ${rankings[i].id} not found`)
                component.errorMessages = ['Something went wrong :(<br />Please try again later']
              }
              return
            }
          }
          component.setScoredEntry(entryId, roundNumber, component.entryNickname, rankings, roundScore)
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
      retrieveScoredEntries: function(entryId, lowerRound, upperRound, foregroundRound) {
        for (let roundNumber = lowerRound; roundNumber <= upperRound; roundNumber++) {
          let isForeground = (roundNumber == foregroundRound)
          this.retrieveScoredEntry(entryId, roundNumber, isForeground)
        }
      },
      scoredEntryExists: function(entryId, roundNumber) {
        if (typeof this.scoredEntries[entryId] == 'undefined') {
          return false
        }
        if (typeof this.scoredEntries[entryId][roundNumber] == 'undefined') {
          return false
        }
        return true
      },
      setScoredEntry: function(entryId, roundNumber, entryNickname, rankings, roundScore) {
        if (typeof this.scoredEntries[entryId] == 'undefined') {
          this.scoredEntries[entryId] = {}
        }
        this.scoredEntries[entryId][roundNumber] = {entryNickname, rankings, roundScore}
      }
    },
    computed: {
      scoredEntryToShow: function() {
        return this.getScoredEntryToShow(this.entryId, this.showRoundNumber)
      },
    },
    watch: {
      entryId: function(newEntryId) {
        this.resetErrorMessages()
        this.showRoundNumber = 0
        let lower = this.roundNumber - preloadBuffer
        let upper = this.roundNumber + preloadBuffer
        this.retrieveScoredEntries(newEntryId, lower, upper, this.roundNumber)
      },
      roundNumber: function(newRoundNumber) {
        this.resetErrorMessages()
        let lower = newRoundNumber - preloadBuffer
        let upper = newRoundNumber + preloadBuffer
        this.retrieveScoredEntries(this.entryId, lower, upper, newRoundNumber)
      },
    }
  }
</script>
