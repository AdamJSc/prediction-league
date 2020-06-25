CREATE TABLE `token` (
    `id` VARCHAR(32) PRIMARY KEY NOT NULL,
    `type` INT(11) NOT NULL,
    `value` VARCHAR(255) NOT NULL,
    `issued_at` DATETIME NOT NULL,
    `expires_at` DATETIME NULL
);
