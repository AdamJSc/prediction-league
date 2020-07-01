CREATE TABLE `scored_entry_prediction` (
    `entry_prediction_id` VARCHAR(36) NOT NULL,
    `standings_id` VARCHAR(36) NOT NULL,
    `rankings` JSON NOT NULL,
    `score` INT(11) NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NULL,
    PRIMARY KEY (entry_prediction_id, standings_id),
    FOREIGN KEY (entry_prediction_id) REFERENCES entry_prediction (id),
    FOREIGN KEY (standings_id) REFERENCES standings (id)
);
