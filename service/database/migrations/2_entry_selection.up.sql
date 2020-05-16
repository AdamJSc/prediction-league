CREATE TABLE `entry_selection` (
    `id` VARCHAR(36) NOT NULL,
    `entry_id` VARCHAR(36) NOT NULL,
    `rankings` JSON NOT NULL,
    `created_at` DATETIME NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (entry_id) REFERENCES entry (id)
);
