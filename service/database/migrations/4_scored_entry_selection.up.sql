CREATE TABLE `scored_entry_selection` (
    `entry_selection_id` VARCHAR(36) NOT NULL,
    `standings_id` VARCHAR(36) NOT NULL,
    `rankings` JSON NOT NULL,
    `score` INT(11) NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NULL,
    PRIMARY KEY (entry_selection_id, standings_id),
    FOREIGN KEY (entry_selection_id) REFERENCES entry_selection (id),
    FOREIGN KEY (standings_id) REFERENCES standings (id)
);
