CREATE TABLE IF NOT EXISTS `season` (
    `id` VARCHAR(10) NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `entries_from` DATETIME NOT NULL,
    `entries_until` DATETIME,
    `start_date` DATETIME NOT NULL,
    `end_date` DATETIME NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME,
    PRIMARY KEY (id)
);
