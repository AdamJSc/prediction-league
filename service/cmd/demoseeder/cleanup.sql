/**
  cleanup all entities created by the demo seeder
 */

# cleanup demo scored entry predictions
DELETE FROM scored_entry_prediction WHERE entry_prediction_id IN (
    SELECT entry_prediction_id FROM entry_prediction WHERE entry_id IN (
        SELECT entry_id FROM entry WHERE season_id="FakeSeason" AND realm_name="localhost"
    )
);

# cleanup demo entry predictions
DELETE FROM entry_prediction WHERE entry_id IN (
    SELECT entry_id FROM entry WHERE season_id="FakeSeason" AND realm_name="localhost"
);

# cleanup demo entries
DELETE FROM entry WHERE season_id="FakeSeason" AND realm_name="localhost";

# cleanup demo standings
DELETE FROM standings WHERE season_id="FakeSeason";
