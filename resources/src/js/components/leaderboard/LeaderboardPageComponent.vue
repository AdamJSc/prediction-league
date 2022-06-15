<template>
  <div class="leaderboard-container">
    <div id="scoredEntryModal" class="modal fade" tabindex="-1" role="dialog">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <button type="button" class="close" data-dismiss="modal" aria-label="Close">
              <span aria-hidden="true">&times;</span>
            </button>
          </div>
          <div class="modal-body">
            <scored-entry
                :entry-id="entryId"
                :entry-nickname="entryNickname"
                :round-number="roundNumber"
                :teams="teams"
                v-on:decrement-round="previousRound"
                v-on:increment-round="nextRound"
            ></scored-entry>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-primary" data-dismiss="modal">Close</button>
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
        v-on:entry-change="changeEntry"
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
        entryId: '', // entry id to retrieve scored entries for
        initialRankings, // rankings belonging to leaderboard of initial round number
        maxRoundNumber: this.initialRoundNumber, // maximum available round number
        roundNumber: this.initialRoundNumber, // round number that is inc/decremented by navigation controls
        teams, // array of team objects with the schema id, name, short_name, crest_url
      }
    },
    methods: {
      changeEntry: function(newEntryId) {
        this.entryId = newEntryId
        $('#scoredEntryModal').modal('show')
      },
      nextRound: function() {
        this.roundNumber++
      },
      previousRound: function() {
        this.roundNumber--
      },
    },
    computed: {
      entryNickname: function() {
        if (this.entryId == '') {
          return ''
        }
        if (typeof this.entries[this.entryId] == 'undefined') {
          return ''
        }
        return this.entries[this.entryId]
      }
    }
  }
</script>
