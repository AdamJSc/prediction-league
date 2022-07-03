CREATE TABLE `mw_result_modifier` (
    `mw_result_id` VARCHAR(36) NOT NULL,
    `order` INT(11) NOT NULL,
    `code` VARCHAR(255) NOT NULL,
    `value` INT(11) NOT NULL,
    PRIMARY KEY (mw_result_id, code),
    FOREIGN KEY (mw_result_id) REFERENCES mw_result (id)
);
