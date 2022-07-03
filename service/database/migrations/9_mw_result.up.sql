CREATE TABLE `mw_result` (
    `id` VARCHAR(36) NOT NULL PRIMARY KEY,
    `mw_submission_id` VARCHAR(36) NOT NULL,
    `team_rankings` JSON NOT NULL,
    `score` INT(11) NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NULL,
    FOREIGN KEY (mw_submission_id) REFERENCES mw_submission (id)
);
