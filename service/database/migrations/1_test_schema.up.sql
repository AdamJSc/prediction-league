CREATE TABLE IF NOT EXISTS `test_table` (
    `id` INT(11) DEFAULT NULL,
    `hello_world` VARCHAR(255) NOT NULL DEFAULT '\"yep\"',
    `is_nice` boolean DEFAULT true,
    PRIMARY KEY (hello_world)
);
