CREATE TABLE IF NOT EXISTS `entry` (
    `id` VARCHAR(36) NOT NULL,
    `lookup_ref` VARCHAR(255) NOT NULL,
    `season_id` VARCHAR(10) NOT NULL,
    `realm_name` VARCHAR(255) NOT NULL,
    `entrant_name` VARCHAR(255) NOT NULL,
    `entrant_nickname` VARCHAR(255) NOT NULL,
    `entrant_email` VARCHAR(255) NOT NULL,
    `status` VARCHAR(255) NOT NULL,
    `payment_method` VARCHAR(255) NULL,
    `payment_ref` VARCHAR(255) NULL,
    `team_id_sequence` JSON NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NULL,
    PRIMARY KEY (id),
    UNIQUE KEY `lookup_ref_index` (lookup_ref),
    UNIQUE KEY `entrant_email_index` (entrant_email, season_id, realm_name),
    UNIQUE KEY `entrant_nickname_index` (entrant_nickname, season_id, realm_name)
);
