CREATE TABLE IF NOT EXISTS `season` (
    `id` VARCHAR(10) NOT NULL,
    `name` VARCHAR(255),
    `entries_from` DATETIME NOT NULL,
    `entries_until` DATETIME,
    `start_date` DATETIME,
    `end_date` DATETIME,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME,
    PRIMARY KEY (id)
);
