<template>
  <div class="scored-entry-prediction-container">
    <transition name="fade">
      <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
        <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
        <p v-for="msg in errorMessages" v-html="msg"></p>
      </div>
    </transition>
    <div v-if="working" class="loader-container">
      <img alt="loader" src="/assets/img/loader-light-bg.svg" />
    </div>
    <div v-if="!working && errorMessages.length == 0">
      <table class="rankings full-width">
        <thead>
        <tr>
          <td colspan="3"></td>
          <td class="text-right text-highlight">Pts</td>
          <td class="text-right text-lolight">Pos</td>
        </tr>
        </thead>
        <tbody>
        <tr v-for="ranking in rankings" class="rankings-row">
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

  export default {
    name: 'ScoredEntryPrediction',
    props: {
      entryId: {
        type: String
      },
      roundNumber: {
        type: Number
      },
      teams: {
        type: Object
      }
    },
    data: function() {
      return {
        working: false,
        errorMessages: [],
        rankings: []
      }
    },
    methods: {
      resetErrorMessages: function() {
        this.errorMessages = []
      },
      refreshScoredEntryPrediction: function(entryId) {
        const vm = this
        vm.working = true
        vm.resetErrorMessages()
        vm.rankings = []
        let url= `/api/entry/${entryId}/scored/${vm.roundNumber}`
        axios.request({
          method: 'get',
          url: url
        }).then(function (response) {
          let rankings = response.data.data.scored.rankings
          for (let i in rankings) {
            if (vm.getTeamByID(rankings[i].id) === null) {
              console.log(`team with id ${rankings[i].id} not found`)
              vm.errorMessages = ['Something went wrong :(<br />Please try again later']
              vm.working = false
              return
            }
          }
          vm.rankings = rankings
          vm.working = false
        }).catch(function (error) {
          console.log(error)
          switch (error.request.status) {
            case 404:
              vm.errorMessages = ["We don't have any data for this round yet!<br />Please try again later"]
              break
            default:
              vm.errorMessages = ['Something went wrong :(<br />Please try again later']
              break
          }
          vm.working = false
        })
      },
      getTeamByID: function(id) {
        for (let i in this.teams) {
          if (this.teams[i].id === id) {
            return this.teams[i]
          }
        }
        return null
      }
    },
    watch: {
      entryId: function(newEntryId) {
        this.refreshScoredEntryPrediction(newEntryId)
      }
    }
  }
</script>
