CREATE TABLE IF NOT EXISTS `standings` (
    `id` VARCHAR(36) NOT NULL,
    `season_id` VARCHAR(10) NOT NULL,
    `round_number` VARCHAR(255) NOT NULL,
    `rankings` JSON NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NULL,
    PRIMARY KEY (id),
    UNIQUE KEY `season_round_index` (season_id, round_number)
);
