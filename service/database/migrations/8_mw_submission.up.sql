CREATE TABLE `mw_submission` (
    `id` VARCHAR(36) NOT NULL,
    `entry_id` VARCHAR(36) NOT NULL,
    `mw_number` INT(11) NOT NULL,
    `team_rankings` JSON NOT NULL,
    `legacy_entry_prediction_id` VARCHAR(36) NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NULL,
    PRIMARY KEY (id), # needed to support mw_result foreign key
    UNIQUE KEY (entry_id, mw_number),
    UNIQUE KEY (legacy_entry_prediction_id, mw_number), # needed to support lookup on these two fields
    FOREIGN KEY (entry_id) REFERENCES entry (id)
);
