<template>
  <div class="leaderboard-container">
    <div id="scoredEntryPredictionModal" class="modal fade" tabindex="-1" role="dialog">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <!--<h5 class="modal-title">{{focusedEntryNickname}} / Score: {{focusedEntryLeaderboardScore}}</h5>-->
            <button type="button" class="close" data-dismiss="modal" aria-label="Close">
              <span aria-hidden="true">&times;</span>
            </button>
          </div>
          <div class="modal-body">
            <!--
            <round-navigation
                :is-working="leaderboardsWorking"
                :max-round-number="maxRoundNumber"
                :round-number="baseRoundNumber"
                v-on:decrement-round="prevRound"
                v-on:increment-round="nextRound"
            ></round-navigation>
            <scored-entry
                :entry-id="focusedEntryId"
                :round-number="focusedRoundNumber"
                :teams="teams"
            ></scored-entry>
            -->
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
          </div>
        </div>
      </div>
    </div>
    <leaderboard
        v-bind:entries="entries"
        v-bind:initial-last-updated-unix="initialLastUpdatedUnix"
        v-bind:initial-rankings="initialRankings"
        v-bind:round-number="roundNumber"
        v-bind:season-id="seasonId"
        v-bind:teams="teams"
        v-on:decrement-round="previousRound"
        v-on:increment-round="nextRound"
    ></leaderboard>
  </div>
</template>

<script>
  export default {
    name: 'LeaderboardPage',
    props: {
      initialLastUpdatedUnix: { // last updated timestamp of leaderboard relating to initial round number
        type: String
      },
      initialRoundNumber: { // round number to activate the state with (always the most recently-active / maximum round number)
        type: Number
      },
      rawEntries: { // json string representing map of entry id to entry nickname
        type: String
      },
      rawRankings: { // json string representing array of ranking objects
        type: String
      },
      rawTeams: { // json string representing array of team objects
        type: String
      },
      seasonId: { // season id relating to current session
        type: String
      }
    },
    data: function() {
      let entries = this.rawEntries === "" ? [] : JSON.parse(this.rawEntries)
      let initialRankings = this.rawRankings === "" ? [] : JSON.parse(this.rawRankings)
      let teams = this.rawTeams === "" ? [] : JSON.parse(this.rawTeams)

      return {
        entries, // map of entry id to entry nickname
        initialRankings, // rankings belonging to leaderboard of initial round number
        roundNumber: this.initialRoundNumber, // round number that is inc/decremented by navigation controls
        maxRoundNumber: this.initialRoundNumber, // maximum available round number
        teams, // array of team objects with the schema id, name, short_name, crest_url
      }
    },
    methods: {
      nextRound: function() {
        this.roundNumber++
      },
      previousRound: function() {
        this.roundNumber--
      },
    }
  }
</script>
